package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录请求耗时和状态码。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		log.Printf("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}
