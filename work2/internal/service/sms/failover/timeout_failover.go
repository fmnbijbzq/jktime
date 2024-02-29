package failover

import (
	"context"
	"example/wb/internal/service/sms"
	"sync/atomic"
)

type TimeOutFailoverSMSService struct {
	svcs []sms.Service
	// 当前正在使用节点
	idx int32
	// 连续几个超时了
	cnt int32
	// 切换阈值, 只读的
	threshold int32
}

func NewTimeOutFailoverSMSService(ss []sms.Service, threshold int32) *TimeOutFailoverSMSService {
	return &TimeOutFailoverSMSService{
		svcs:      ss,
		idx:       0,
		cnt:       10,
		threshold: threshold,
	}
}

func (t *TimeOutFailoverSMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)

	// 超过了阈值，执行切换操作
	if cnt >= t.threshold {
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			// 重置这个计数
			atomic.StoreInt32(&t.cnt, 0)
		}
		idx = newIdx
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case nil:
		// 连续超时
		atomic.StoreInt32(&t.cnt, 0)
		return nil
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	default:
		// 遇到了错误，但是又不是超时错误，这个时候，你要考虑怎么搞
		// 我可以增加，也可以不增加
		// 如果强调一定是超时，那么就不增加
		// 如果是EOF之类的错误，你还可以考虑直接切换
	}
	return err
}
