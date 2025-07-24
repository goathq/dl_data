package models

import (
	"time"

	"gorm.io/gorm"
)

// 用户表
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"uniqueIndex;size:64"` // 交易所用户名
	Email     string `gorm:"size:128"`            // 用户邮箱
	CreatedAt time.Time
	UpdatedAt time.Time
}

// 资产表
type Asset struct {
	ID        uint   `gorm:"primaryKey"`
	Symbol    string `gorm:"uniqueIndex;size:32"` // 资产符号，如 BTC, ETH
	Name      string `gorm:"size:64"`             // 资产名称
	CreatedAt time.Time
}

// 用户资产余额表
type UserAsset struct {
	ID        uint    `gorm:"primaryKey"`
	UserID    uint    `gorm:"index:idx_user_asset,unique"` // 用户ID
	AssetID   uint    `gorm:"index:idx_user_asset,unique"` // 资产ID
	Balance   float64 `gorm:"type:decimal(36,18)"`         // 可用余额
	Locked    float64 `gorm:"type:decimal(36,18)"`         // 锁定余额(如质押中)
	CreatedAt time.Time
	UpdatedAt time.Time

	User  User  `gorm:"foreignKey:UserID"`
	Asset Asset `gorm:"foreignKey:AssetID"`
}

// LaunchPool项目表
type LaunchPool struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100"`            // 项目名称
	StakeAsset  string    `gorm:"size:32;index"`       // 质押资产符号
	RewardAsset string    `gorm:"size:32;index"`       // 奖励资产符号
	StartTime   time.Time `gorm:"index"`               // 开始时间
	EndTime     time.Time `gorm:"index"`               // 结束时间
	APY         float64   `gorm:"type:decimal(10,4)"`  // 年化收益率
	TotalStaked float64   `gorm:"type:decimal(36,18)"` // 总质押量
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// 用户质押记录表
type UserStake struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index:idx_user_pool"` // 用户ID
	PoolID      uint      `gorm:"index:idx_user_pool"` // 池子ID
	Amount      float64   `gorm:"type:decimal(36,18)"` // 质押数量
	Reward      float64   `gorm:"type:decimal(36,18)"` // 累计奖励
	StakedAt    time.Time `gorm:"index"`               // 质押时间
	LastClaimAt time.Time // 上次领取时间
	CreatedAt   time.Time
	UpdatedAt   time.Time

	User User       `gorm:"foreignKey:UserID"`
	Pool LaunchPool `gorm:"foreignKey:PoolID"`
}

// 资产交易记录表
type AssetTransaction struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      uint      `gorm:"index"`               // 用户ID
	AssetSymbol string    `gorm:"size:32;index"`       // 资产符号
	Amount      float64   `gorm:"type:decimal(36,18)"` // 数量
	Type        string    `gorm:"size:32;index"`       // 类型: deposit/withdraw/stake/unstake/reward
	Status      string    `gorm:"size:32;index"`       // 状态: pending/completed/failed
	TxID        string    `gorm:"size:128;index"`      // 交易ID
	CreatedAt   time.Time `gorm:"index"`
	UpdatedAt   time.Time

	User User `gorm:"foreignKey:UserID"`
}

// MigrateDB 自动迁移数据库
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Asset{},
		&UserAsset{},
		&LaunchPool{},
		&UserStake{},
		&AssetTransaction{},
	)
}
