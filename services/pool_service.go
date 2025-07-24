package services

import (
	"errors"
	"math/rand"
	"time"

	"github.com/goathq/dl_data/models"

	"gorm.io/gorm"
)

type PoolService struct {
	db *gorm.DB
}

func NewPoolService(db *gorm.DB) *PoolService {
	return &PoolService{db: db}
}

// 用户质押
func (s *PoolService) Stake(userID uint, poolID uint, AssetID uint, amount float64) error {
	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查池子是否存在且有效
	var pool models.LaunchPool
	if err := tx.First(&pool, poolID).Error; err != nil {
		tx.Rollback()
		return errors.New("pool not found")
	}

	now := time.Now()
	if now.Before(pool.StartTime) || now.After(pool.EndTime) {
		tx.Rollback()
		return errors.New("pool not active")
	}

	// 2. 检查用户资产是否足够
	var userAsset models.UserAsset
	if err := tx.Where("user_id = ? AND asset_id IN (SELECT id FROM assets WHERE symbol = ?)",
		userID, pool.StakeAsset).First(&userAsset).Error; err != nil {
		tx.Rollback()
		return errors.New("asset not found")
	}

	if userAsset.Balance < amount {
		tx.Rollback()
		return errors.New("insufficient balance")
	}

	// 3. 扣除用户可用余额并增加锁定余额
	if err := tx.Model(&userAsset).Updates(map[string]interface{}{
		"balance": gorm.Expr("balance - ?", amount),
		"locked":  gorm.Expr("locked + ?", amount),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. 创建质押记录
	stake := models.UserStake{
		UserID:      userID,
		PoolID:      poolID,
		Amount:      amount,
		Reward:      0,
		StakedAt:    now,
		LastClaimAt: now,
	}

	if err := tx.Create(&stake).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 5. 更新池子总质押量
	if err := tx.Model(&pool).Update("total_staked", gorm.Expr("total_staked + ?", amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 6. 创建交易记录
	txRecord := models.AssetTransaction{
		UserID:      userID,
		AssetSymbol: pool.StakeAsset,
		Amount:      amount,
		Type:        "stake",
		Status:      "completed",
		TxID:        generateTxID(),
	}

	if err := tx.Create(&txRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 用户赎回
func (s *PoolService) Unstake(userID uint, poolID uint, amount float64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查质押记录
	var stake models.UserStake
	if err := tx.Where("user_id = ? AND pool_id = ?", userID, poolID).First(&stake).Error; err != nil {
		tx.Rollback()
		return errors.New("stake record not found")
	}

	if stake.Amount < amount {
		tx.Rollback()
		return errors.New("unstake amount exceeds staked amount")
	}

	// 2. 获取池子信息
	var pool models.LaunchPool
	if err := tx.First(&pool, poolID).Error; err != nil {
		tx.Rollback()
		return errors.New("pool not found")
	}

	// 3. 计算并发放奖励
	reward := s.calculateReward(&stake, &pool)
	if reward > 0 {
		if err := s.distributeReward(tx, userID, pool.RewardAsset, reward); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 4. 更新用户资产
	var userAsset models.UserAsset
	if err := tx.Where("user_id = ? AND asset_id IN (SELECT id FROM assets WHERE symbol = ?)",
		userID, pool.StakeAsset).First(&userAsset).Error; err != nil {
		tx.Rollback()
		return errors.New("asset not found")
	}

	if err := tx.Model(&userAsset).Updates(map[string]interface{}{
		"balance": gorm.Expr("balance + ?", amount),
		"locked":  gorm.Expr("locked - ?", amount),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 5. 更新质押记录
	if stake.Amount == amount {
		// 全部赎回，删除记录
		if err := tx.Delete(&stake).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 部分赎回，更新记录
		if err := tx.Model(&stake).Updates(map[string]interface{}{
			"amount":        gorm.Expr("amount - ?", amount),
			"reward":        0, // 奖励已发放，重置为0
			"last_claim_at": time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 6. 更新池子总质押量
	if err := tx.Model(&pool).Update("total_staked", gorm.Expr("total_staked - ?", amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 7. 创建交易记录
	txRecord := models.AssetTransaction{
		UserID:      userID,
		AssetSymbol: pool.StakeAsset,
		Amount:      amount,
		Type:        "unstake",
		Status:      "completed",
		TxID:        generateTxID(),
	}

	if err := tx.Create(&txRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 领取奖励
func (s *PoolService) ClaimReward(userID uint, poolID uint) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查质押记录
	var stake models.UserStake
	if err := tx.Where("user_id = ? AND pool_id = ?", userID, poolID).First(&stake).Error; err != nil {
		tx.Rollback()
		return errors.New("stake record not found")
	}

	// 2. 获取池子信息
	var pool models.LaunchPool
	if err := tx.First(&pool, poolID).Error; err != nil {
		tx.Rollback()
		return errors.New("pool not found")
	}

	// 3. 计算奖励
	reward := s.calculateReward(&stake, &pool)
	if reward <= 0 {
		tx.Rollback()
		return errors.New("no reward to claim")
	}

	// 4. 发放奖励
	if err := s.distributeReward(tx, userID, pool.RewardAsset, reward); err != nil {
		tx.Rollback()
		return err
	}

	// 5. 更新质押记录
	if err := tx.Model(&stake).Updates(map[string]interface{}{
		"reward":        0, // 奖励已发放，重置为0
		"last_claim_at": time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 6. 创建交易记录
	txRecord := models.AssetTransaction{
		UserID:      userID,
		AssetSymbol: pool.RewardAsset,
		Amount:      reward,
		Type:        "reward",
		Status:      "completed",
		TxID:        generateTxID(),
	}

	if err := tx.Create(&txRecord).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// 计算奖励
func (s *PoolService) calculateReward(stake *models.UserStake, pool *models.LaunchPool) float64 {
	now := time.Now()
	if now.After(pool.EndTime) {
		now = pool.EndTime
	}

	// 简单计算: (质押金额 * APY * 时间比例)
	hours := now.Sub(stake.LastClaimAt).Hours()
	days := hours / 24
	years := days / 365

	reward := stake.Amount * pool.APY * years
	return reward
}

// 发放奖励
func (s *PoolService) distributeReward(tx *gorm.DB, userID uint, assetSymbol string, amount float64) error {
	// 1. 获取或创建用户资产记录
	var userAsset models.UserAsset
	err := tx.Where("user_id = ? AND asset_id IN (SELECT id FROM assets WHERE symbol = ?)",
		userID, assetSymbol).First(&userAsset).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果用户没有该资产，先创建记录
			var asset models.Asset
			if err := tx.Where("symbol = ?", assetSymbol).First(&asset).Error; err != nil {
				return errors.New("asset not found")
			}

			userAsset = models.UserAsset{
				UserID:  userID,
				AssetID: asset.ID,
				Balance: amount,
				Locked:  0,
			}

			if err := tx.Create(&userAsset).Error; err != nil {
				return err
			}
			return nil
		}
		return err
	}

	// 2. 更新用户资产余额
	if err := tx.Model(&userAsset).Update("balance", gorm.Expr("balance + ?", amount)).Error; err != nil {
		return err
	}

	return nil
}

func generateTxID() string {
	// 实现一个生成唯一交易ID的方法
	return "tx_" + time.Now().Format("20060102150405") + "_" + randStr(8)
}

func randStr(n int) string {
	// 实现一个生成随机字符串的方法
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
