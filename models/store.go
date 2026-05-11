package models

import "time"

// Struct Store ini akan diubah GORM menjadi tabel 'stores'
type Store struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NamaToko     string    `gorm:"type:varchar(100);not null" json:"nama_toko"`
	Alamat       string    `gorm:"type:text" json:"alamat"`
	Telepon      string    `gorm:"type:varchar(20)" json:"telepon"`
	BusinessType string    `gorm:"type:varchar(50);default:'kelontong'" json:"business_type"` // kelontong, cafe, laundry
	PajakPersen  float64   `gorm:"type:decimal(5,2);default:0" json:"pajak_persen"` // Isi 0 untuk warung, 11/12 untuk Cafe
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}