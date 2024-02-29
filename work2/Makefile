.PHONY: docker
docker:
	@rm webook || true
	@go mod tidy
	@GOOS=linux GOARCH=arm go build -tags=k8s -o webook .
	@docker rmi -f bzq/webook:v0.0.1
	@docker build -t bzq/webook:v0.0.1 .

.PHONY: mock
mock:
	@mockgen -source=internal/service/user.go -package=svcmock -destination=internal/service/mocks/user_mock.go
	@mockgen -source=internal/service/code.go -package=svcmock -destination=internal/service/mocks/code_mock.go
	@mockgen -source=internal/repository/user.go -destination=internal/repository/mock/user_mock.go -package=repomock
	@mockgen -source=internal/repository/code.go -destination=internal/repository/mock/code_mock.go -package=repomock
	@mockgen -source=internal/repository/async_sms.go -destination=internal/repository/mock/sms_mock.go -package=repomock
	@mockgen -source=internal/repository/dao/user.go -destination=internal/repository/dao/mock/user_mock.go -package=daomock
	@mockgen -source=internal/repository/dao/async_sms.go -destination=internal/repository/dao/mock/sms_mock.go -package=daomock
	@mockgen -source=internal/repository/cache/code.go -destination=internal/repository/cache/mock/code_mock.go -package=cachemock
	@mockgen -source=internal/repository/cache/user.go -destination=internal/repository/cache/mock/user_mock.go -package=cachemock
	@mockgen -source=internal/service/sms/types.go -package=smsmock -destination=internal/service/sms/mocks/sms_mock.go 
	@mockgen -source=pkg/limiter/types.go -package=limitmock -destination=pkg/limiter/mocks/limiter_mock.go 
	@mockgen -package=redismock -destination=internal/repository/cache/redismock/cmd_mock.go github.com/redis/go-redis/v9 Cmdable 
	@go mod tidy