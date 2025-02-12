package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	Status   int           `json:"status"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
}

type LoggerMiddlewareBuilder struct {
	logFunc       func(ctx context.Context, info *AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLoggerMiddlewareBuilder(logF func(ctx context.Context, info *AccessLog)) *LoggerMiddlewareBuilder {
	return &LoggerMiddlewareBuilder{
		logFunc: logF,
	}
}

func (b *LoggerMiddlewareBuilder) WithReqBody() *LoggerMiddlewareBuilder {
	b.allowReqBody = true
	return b
}

func (b *LoggerMiddlewareBuilder) WithRespBody() *LoggerMiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *LoggerMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if len(path) > 1024 {
			path = path[:1024]
		}
		aLogInfo := &AccessLog{
			Path:   path,
			Method: ctx.Request.Method,
		}
		if b.allowReqBody {
			bodyBytes, _ := ctx.GetRawData()
			if len(bodyBytes) > 2048 {
				aLogInfo.ReqBody = string(bodyBytes[:2048])
			} else {
				aLogInfo.ReqBody = string(bodyBytes)
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		start := time.Now()
		if b.allowRespBody {
			ctx.Writer = &loggerResponseWriter{ctx.Writer, aLogInfo}
		}
		// 执行下一个中间件
		ctx.Next()
		// 记录日志
		aLogInfo.Duration = time.Since(start)
		aLogInfo.Status = ctx.Writer.Status()
		defer func() {
			b.logFunc(context.Background(), aLogInfo)
		}()
	}
}

type loggerResponseWriter struct {
	gin.ResponseWriter
	acLog *AccessLog
}

func (w *loggerResponseWriter) Write(b []byte) (int, error) {
	w.acLog.RespBody = string(b)
	return w.ResponseWriter.Write(b)
}

func (w *loggerResponseWriter) WriteString(s string) (int, error) {
	w.acLog.RespBody = s
	return w.ResponseWriter.WriteString(s)
}

func (w *loggerResponseWriter) WriteHeader(code int) {
	w.acLog.Status = code
	w.ResponseWriter.WriteHeader(code)
}
