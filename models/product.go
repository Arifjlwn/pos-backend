package models

import "time"

type Product struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	StoreID    uint      `gorm:"not null" json:"store_id"` // Kunci agar produk nggak nyasar ke toko lain
	SKU        *string   `gorm:"type:varchar(50);uniqueIndex" json:"sku"` // Barcode/Kode Barang (Pakai * biar bisa NULL buat Cafe/Jasa)
	NamaProduk string    `gorm:"type:varchar(150);not null" json:"nama_produk"`
	Kategori   string    `gorm:"type:varchar(50)" json:"kategori"`
	HargaModal float64   `gorm:"type:decimal(10,2);default:0" json:"harga_modal"`
	HargaJual  float64   `gorm:"type:decimal(10,2);not null" json:"harga_jual"`
	Stok       int       `gorm:"default:0" json:"stok"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relasi ke Store
	Store Store `gorm:"foreignKey:StoreID" json:"-"`
}