package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户数据模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Name      string         `json:"name" gorm:"not null;size:100"`      // 用户姓名
	Email     string         `json:"email" gorm:"not null;unique;size:100"` // 用户邮箱
	Age       int            `json:"age"`                                  // 用户年龄
	Status    int            `json:"status" gorm:"default:1"`              // 用户状态 1-正常 0-禁用
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}