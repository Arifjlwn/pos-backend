package models

import "time"

// Struct Store ini akan diubah GORM menjadi tabel 'stores'
type Store struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NamaToko     string    `gorm:"type:varchar(100);not null" json:"nama_toko"`
	Alamat       string    `gorm:"type:text" json:"alamat"`
	Telepon      string    `gorm:"type:varchar(20)" json:"telepon"`
	BusinessType string    `gorm:"type:varchar(50);default:'kelontong'" json:"business_type"` // kelontong, cafe, laundry
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}