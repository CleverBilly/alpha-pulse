package utils

import "github.com/gin-gonic/gin"

// Success 统一成功响应结构。
func Success(data any) gin.H {
	return gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	}
}

// Error 统一错误响应结构。
func Error(code int, message string) gin.H {
	return gin.H{
		"code":    code,
		"message": message,
	}
}
