package model

import (
	"gorm.io/gorm"
)

type KV struct {
    gorm.Model
    Key   string `gorm:"column:key_name;size:255;uniqueIndex;not null"`
    Value string `gorm:"not null"`
}