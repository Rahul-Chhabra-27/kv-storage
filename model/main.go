package model


type KV struct {
    ID    uint   `gorm:"primaryKey"`
    Key   string `gorm:"column:key_name;size:255;uniqueIndex;not null"`
    Value string `gorm:"not null"`
}