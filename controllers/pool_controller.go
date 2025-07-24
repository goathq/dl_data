package controllers

import (
	"net/http"

	"github.com/goathq/dl_data/models"
	"github.com/goathq/dl_data/services"
	"github.com/goathq/dl_data/utils"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type PoolController struct {
	db          *gorm.DB
	poolService *services.PoolService
}

func NewPoolController(db *gorm.DB, poolService *services.PoolService) *PoolController {
	return &PoolController{
		db:          db,
		poolService: poolService}
}

// Stake 用户质押
func (c *PoolController) Stake(ctx *gin.Context) {
	var req struct {
		UserID  uint    `json:"user_id" binding:"required"`
		PoolID  uint    `json:"pool_id" binding:"required"`
		AssetID uint    `json:"asset_id" binding:"required"`
		Amount  float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.poolService.Stake(req.UserID, req.PoolID, req.AssetID, req.Amount); err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(ctx, gin.H{
		"message": "stake successful",
	})
}

// Unstake 用户赎回
func (c *PoolController) Unstake(ctx *gin.Context) {
	var req struct {
		UserID uint    `json:"user_id" binding:"required"`
		PoolID uint    `json:"pool_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.poolService.Unstake(req.UserID, req.PoolID, req.Amount); err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(ctx, gin.H{
		"message": "unstake successful",
	})
}

// ClaimReward 领取奖励
func (c *PoolController) ClaimReward(ctx *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
		PoolID uint `json:"pool_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.poolService.ClaimReward(req.UserID, req.PoolID); err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(ctx, gin.H{
		"message": "reward claimed successfully",
	})
}

// GetUserStakes 获取用户质押记录
func (c *PoolController) GetUserStakes(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	if userID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "user_id is required")
		return
	}

	var stakes []models.UserStake
	if err := c.db.Where("user_id = ?", userID).Preload("Pool").Find(&stakes).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(ctx, gin.H{
		"stakes": stakes,
	})
}

// GetPoolInfo 获取池子信息
func (c *PoolController) GetPoolInfo(ctx *gin.Context) {
	poolID := ctx.Query("pool_id")
	if poolID == "" {
		utils.RespondWithError(ctx, http.StatusBadRequest, "pool_id is required")
		return
	}

	var pool models.LaunchPool
	if err := c.db.First(&pool, poolID).Error; err != nil {
		utils.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithSuccess(ctx, gin.H{
		"pool": pool,
	})
}
