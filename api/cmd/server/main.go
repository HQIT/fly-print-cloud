package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin路由
	r := gin.Default()

	// 基础路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "fly-print-cloud-api",
		})
	})

	// 启动服务器
	log.Println("Starting fly-print-cloud API server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
