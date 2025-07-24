package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string
	Password string
	ELO      int `gorm:"default:1000"`
}
