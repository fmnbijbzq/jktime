package domain

type Sms struct {
	ID        int
	Phone     string
	Code      string
	TplId     string
	Biz       string
	CreatedAt int64
}

type AsyncSms struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
	// 重试的配置
	RetryMax int
}
