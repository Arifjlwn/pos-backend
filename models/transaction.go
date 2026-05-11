package models

import "time"

// 1. Model untuk Kepala Struk (Header)
type Transaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StoreID      uint      `gorm:"not null" json:"store_id"`
	UserID       uint      `gorm:"not null" json:"user_id"` // ID Kasir atau Owner yang input
	NoInvoice    string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"no_invoice"` // Contoh: INV-20260511-0001

	SubTotal     float64   `gorm:"type:decimal(10,2);not null" json:"sub_total"`  // Total murni harga barang
	Pajak        float64   `gorm:"type:decimal(10,2);default:0" json:"pajak"`     // Nominal rupiah dari PPN (jika ada)
	Pembulatan   float64   `gorm:"type:decimal(10,2);default:0" json:"pembulatan"`// Selisih pembulatan (contoh: -15)
	TotalHarga   float64   `gorm:"type:decimal(10,2);not null" json:"total_harga"`// Final yang ditagih ke pelanggan (contoh: 21300)
	
	NominalBayar float64   `gorm:"type:decimal(10,2);not null" json:"nominal_bayar"`
	Kembalian    float64   `gorm:"type:decimal(10,2);not null" json:"kembalian"`
	CreatedAt    time.Time `json:"created_at"`

	// Relasi ke tabel rincian barang
	Details []TransactionDetail `gorm:"foreignKey:TransactionID" json:"details"`
}

// 2. Model untuk Rincian Barang di Struk (Body)
type TransactionDetail struct {
	ID            uint    `gorm:"primaryKey" json:"id"`
	TransactionID uint    `gorm:"not null" json:"transaction_id"`
	ProductID     uint    `gorm:"not null" json:"product_id"`
	HargaSatuan   float64 `gorm:"type:decimal(10,2);not null" json:"harga_satuan"`
	Kuantitas     int     `gorm:"not null" json:"kuantitas"`
	SubTotal      float64 `gorm:"type:decimal(10,2);not null" json:"sub_total"`

	// Relasi untuk narik nama produk (opsional, buat nampilin di nota)
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}