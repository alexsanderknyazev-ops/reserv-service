package models

import (
	"time"
)

type ItemType string

const (
	ItemTypeWeapon ItemType = "weapon"
	ItemTypeArmor  ItemType = "armor"
	ItemTypeOther  ItemType = "other"
)

type Reserve struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	IdItem    int64     `gorm:"type:bigint;not null;index" json:"id_item"`
	IdUser    int64     `gorm:"type:bigint;not null;index" json:"id_user"`
	ItemType  ItemType  `gorm:"type:varchar(20);not null;default:'weapon'" json:"item_type"`
	Status    string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Reserve) TableName() string {
	return "reserve"
}

// Request структуры
type ReserveRequest struct {
	IdUser   int64    `json:"id_user" binding:"required,min=1"`
	IdItem   int64    `json:"id_item" binding:"required,min=1"`
	ItemType ItemType `json:"item_type" binding:"required,oneof=weapon armor other"`
}

type ReserveResponse struct {
	ID        int64     `json:"id"`
	IdItem    int64     `json:"id_item"`
	IdUser    int64     `json:"id_user"`
	ItemType  string    `json:"item_type"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
