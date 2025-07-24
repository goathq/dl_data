package main

import (
	"log"

	"github.com/goathq/dl_data/config"
	"github.com/goathq/dl_data/models"
	"github.com/goathq/dl_data/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.InitializeDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 自动迁移数据库
	if err := models.MigrateDB(config.AppConfig.DB); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化Gin
	router := gin.Default()

	// 设置路由
	routes.SetupRoutes(config.AppConfig.DB, router)

	// 启动服务器
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
