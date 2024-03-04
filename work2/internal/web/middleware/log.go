package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type LogMiddlewareBuilder struct {
	logFn         func(ctx context.Context, l AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFn func(ctx context.Context, l AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFn: logFn,
	}

}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 防止黑客攻击，path过长
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := ctx.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			// body是一个stream流，需要复制一个ctx副本或者再把body放回去
			body, _ := ctx.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		start := time.Now()
		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}

		}

		defer func() {
			al.Duration = time.Since(start)
			l.logFn(ctx, al)
		}()
		ctx.Next()

	}

}

type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.al.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	RespBody string        `json:"resp_body"`
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration"`
}
