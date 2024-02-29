package async

import (
	"context"
	"example/wb/internal/domain"
	"example/wb/internal/repository"
	"example/wb/internal/service/sms"
	"example/wb/internal/service/sms/ratelimit"
	"log"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Service struct {
	svc      sms.Service
	repo     repository.AsyncSmsRepository
	durTimes []int
	mu       sync.Mutex
}

// type AsyncSMSService struct {
// 	svc      sms.Service
// 	smsRepo  repository.SmsRepository
// 	codeRepo repository.CodeRepository
// 	dur      time.Duration // 重试间隔
// 	times    int           // 重试次数
// 	mu       sync.Mutex
// }

func NewService(svc sms.Service,
	repo repository.AsyncSmsRepository) *Service {
	res := &Service{
		svc:  svc,
		repo: repo,
	}
	go func() {
		res.StartAsyncCycle()
	}()
	return res
}

// 原理：抢占式调度
func (s *Service) StartAsyncCycle() {
	// 防止测试时，偶发性的失败（原理未知）
	time.Sleep(time.Second * 3)
	for {
		s.AsyncSend()
	}

}

func (s *Service) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch err {
	case nil:
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.svc.Send(ctx, as.TplId, as.Args, as.Numbers...)

		if err != nil {
			log.Printf("执行异步发送短信失败, err: %s, id: %d", err, as.Id)
		}
		res := err == nil

		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			log.Printf("执行异步发送短信成功,但是数据库标记失败 err: %s, id: %d", err, as.Id)
		}
	case repository.ErrWaitingSMSNotFound:
		// 数据库里面没有发送失败的消息，可以考虑自由设置休息时间
		time.Sleep(time.Second * 5)
	default:
		log.Printf("抢占异步发送短信失败, err: %s", err)
		time.Sleep(time.Second * 5)
	}

}

func (s *Service) needAsync(curTime int, threshold int) bool {
	var mean int
	s.mu.Lock()
	s.durTimes = append(s.durTimes, curTime)
	if len(s.durTimes) >= 3 {
		var sum int
		for _, val := range s.durTimes {
			sum += val
		}
		mean = sum / len(s.durTimes)
		s.durTimes = append([]int{}, s.durTimes[1:]...)
	} else {
		mean = curTime
	}
	s.mu.Unlock()
	return mean >= threshold
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now().UnixMilli()
	err := s.svc.Send(ctx, tplId, args, numbers...)
	end := time.Now().UnixMilli()
	durTime := int(end - start)

	if err == ratelimit.ErrSMSLimitRate || s.needAsync(durTime, 500) {
		return s.repo.Create(ctx,
			domain.AsyncSms{
				TplId:    tplId,
				Args:     args,
				Numbers:  numbers,
				RetryMax: 3,
			},
		)

	}
	return err
}

// func (s *Service) needAsync() bool {
// 	return true
// }

// 1. 应该创建一个数据库表用来存储限流，崩溃后的数据
// 1.1 通过gorm进行异步请求表的搭建
// 2. 数据库字段应该包括 tplId, args(也就是验证码) numbers(电话号码)
// 3. 创建AsyncSend异步方法, 该方法通过gorm读取Send所需参数根据dur和times进行异步发送
// 4. 在Send方法中使用go func()的形式创建异步方法
// 5. 判定服务商已经崩溃
//		1. 通过谷歌bing查询网络上的相关帖子，查找成熟的判断方式
//		1.1 如果没有可以查找短信服务上有哪些返回异常可以辅助判断
//      1.2 轮询所有服务商，如果没有一个成功那么判定服务商已经崩溃
//		2. 通过chatgpt询问判定服务上崩溃的方法
// 6. 缺点分析(可以从继续细化功能的不足之处，以及用户体验，对数据库的压力太大, 本地网络出错可能会造成数据丢失)
// 适合场景：
// 	1.适合不需要及时响应且需要保证短信一定到收到的场景
// 优点:
//  1.极大的提高了短信服务的可靠性，能够极大的保证短信服务的正常运行
//  2.即使面对短信服务商崩溃的情况，也能依靠本地数据库，在一定的延迟下，进行重发
// 缺点:
//  1.如果短信进入同步转异步中，短信的发送会有一定的延迟
//  2.同时会给数据库带来一定的负载
//  3.如果本地数据库服务宕机的过程中，会造成部分短信的丢失
//  4.由于缺少biz字段，目前短信重发仅能针对登录服务做同步转移步
// 改进方案：
//  针对1不足，由于设定有重试次数和时间间隔，可以根据业务需求设计合适的时间间隔
//  针对2不足，可以考虑将没有发送的信息存入redis中，来减少数据库的压力
//  针对3不足，可以增加本地缓存，在数据库（redis）崩溃的时候，将数据存入本地缓存中
//  针对4不足，可以考虑修改部分代码，将biz字段传过来

// func (s *AsyncSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
// 	err := s.svc.Send(ctx, tplId, args, numbers...)
// 	if err == nil {
// 		return err
// 	}
// 	// 触发限流，发送到数据库中, 这个装饰器应该放到限流之后
// 	if err == ratelimit.ErrSMSLimitRate {
// 		for i := 0; i < len(args); i++ {
// 			err := s.Save(ctx, tplId, args[i], numbers[i])
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		}
// 		return ratelimit.ErrSMSLimitRate
// 	} else {
// 		// 尝试ping服务商的域名, 如果服务商的崩溃大概率是服务器宕机了, 如果ping不通
// 		// 那么证明服务商崩溃了
// 		// 后续改进，需要将不同运营商的短信服务地址抽取出来，方便动态获取
// 		// ok, err := s.Ping(ctx, "sms.tencentcloudapi.com")
// 		ok, err := s.Ping(ctx, "adsfas")
// 		if !ok || err != nil {
// 			for i := 0; i < len(args); i++ {
// 				err := s.Save(ctx, tplId, args[i], numbers[i])
// 				if err != nil {
// 					log.Println(err)
// 				}
// 			}
// 		}
// 	}

// 	// for _, phone := range args {
// 	// 	go s.AsyncSend(ctx, phone)
// 	// }

// 	return err
// }

// func (s *AsyncSMSService) Save(ctx context.Context, tplId string, args string, number string) error {
// 	err := s.smsRepo.Create(ctx, domain.Sms{
// 		Phone: args,
// 		TplId: tplId,
// 		Code:  number,
// 		Biz:   bizLogin,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *AsyncSMSService) AsyncSendV1() {

// }

// func (s *AsyncSMSService) AsyncSend(ctx context.Context, phone string) error {
// 	s.mu.Lock()
// 	ss, err := s.smsRepo.FindByPhone(ctx, phone)
// 	if err != nil {
// 		return err
// 	}
// 	for _, ds := range ss {
// 		err = s.smsRepo.DeleteById(ctx, ds.ID)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	s.mu.Unlock()

// 	for _, ds := range ss {
// 		// 重试间隔
// 		time.Sleep(s.dur)
// 		for i := 0; i < s.times; i++ {
// 			// 考虑到如果进入异步发送那么意味着本地redis运行正常
// 			// 短信运营商大概率有问题，那么先进行短信的发送
// 			// 然后进行redis的缓存
// 			err := s.svc.Send(ctx, ds.TplId, []string{ds.Phone}, ds.Code)
// 			if err != nil {
// 				log.Printf("第%d次发送失败，原因为%s", i+1, err)
// 			} else {
// 				err = s.codeRepo.Set(ctx, ds.Biz, ds.Phone, ds.Code)
// 				if err != nil {
// 					log.Printf("第%d次缓存失败，原因为%s", i, err)
// 				}
// 				break
// 			}
// 		}
// 	}
// 	return nil
// }

func (s *Service) Ping(ctx context.Context, endpoint string) (bool, error) {
	pinger, err := probing.NewPinger(endpoint)
	if err != nil {
		return false, err
	}
	pinger.Count = 3
	err = pinger.Run()
	if err != nil {
		return false, err
	}
	return true, nil

}
