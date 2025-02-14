package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webok/pkg/ginx"
	"webok/pkg/logger"
)

type ErrorLoggerBuilder struct {
	l logger.Logger
}

func NewErrorLoggerBuilder(l logger.Logger) *ErrorLoggerBuilder {
	return &ErrorLoggerBuilder{l: l}
}

func (m *ErrorLoggerBuilder) Build() gin.HandlerFunc {
	if m.l == nil {
		panic("ErrorLoggerBuilder:logger is nil")
	}
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {

				m.l.Error(c.Request.URL.Path, logger.Field{
					Key: "err",
					Val: err.Err,
				})
			}
			// 统一返回错误响应
			c.JSON(http.StatusOK, ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			})
		}
	}

}
