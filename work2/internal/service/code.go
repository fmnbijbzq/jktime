package service

import (
	"context"
	"example/wb/internal/repository"
	"example/wb/internal/service/sms"
	"fmt"
	"math/rand"
)

var ErrCodeVertifyTooMany = repository.ErrCodeVertifyTooMany
var ErrSendTooMany = repository.ErrSendTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Vertify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, sms sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  sms,
	}

}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 短信的模板id，一般为常量，不进行修改
	const tplId = "1263395"
	return svc.sms.Send(ctx, tplId, []string{code}, phone)

}
func (svc *codeService) Vertify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Vertify(ctx, biz, phone, inputCode)
	if err == repository.ErrCodeVertifyTooMany {
		// 屏蔽验证次数过多的错误
		return false, nil
	}
	if !ok || err != nil {
		return false, err
	}
	return true, nil
}

func (svc *codeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)

}
