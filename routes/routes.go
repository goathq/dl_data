package routes

import (
	"github.com/goathq/dl_data/controllers"
	"github.com/goathq/dl_data/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, router *gin.Engine) {
	// 初始化服务
	poolService := services.NewPoolService(db)

	// 初始化控制器
	poolController := controllers.NewPoolController(db, poolService)

	// API路由
	api := router.Group("/api/v1")
	{
		// LaunchPool相关接口
		pool := api.Group("/pool")
		{
			pool.POST("/stake", poolController.Stake)
			pool.POST("/unstake", poolController.Unstake)
			pool.POST("/claim", poolController.ClaimReward)
			pool.GET("/stakes", poolController.GetUserStakes)
			pool.GET("/info", poolController.GetPoolInfo)
		}
	}
}
