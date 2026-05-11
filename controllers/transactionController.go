package controllers

import (
	"fmt"
	"math"
	"net/http"
	"pos-backend/config"
	"pos-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Wadah untuk menangkap keranjang belanjaan dari Frontend
type CartItem struct {
	ProductID uint `json:"product_id" binding:"required"`
	Kuantitas int  `json:"kuantitas" binding:"required,gt=0"`
}

// Wadah utama yang dikirim Frontend saat tombol "Bayar" diklik
type TransactionInput struct {
	Items        []CartItem `json:"items" binding:"required,gt=0"`
	NominalBayar float64    `json:"nominal_bayar" binding:"required"`
}

func CreateTransaction(c *gin.Context) {
	storeID, _ := c.Get("store_id")
	userID, _ := c.Get("user_id")

	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format keranjang tidak sesuai!"})
	}

	var savedTransaction models.Transaction

	// --- MULAI DATABASE TRANSACTION (Biar Aman dari Error) ---
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil data Toko untuk ngecek PPN-nya
		var store models.Store
		if err := tx.First(&store, storeID).Error; err != nil {
			return err
		}

		var subTotal float64
		var details []models.TransactionDetail

		// 2. Scan satu per satu barang di keranjang
		for _, item := range input.Items {
			var product models.Product

			// Cek barang nya ada nga, dan pastikan milik toko ini
			if err := tx.Where("id = ? AND store_id = ?", item.ProductID, storeID).First(&product).Error; err != nil {
				return fmt.Errorf("barang dengan ID %d tidak ditemukan", item.ProductID)
			}

			// Cek Stok
			if product.Stok < item.Kuantitas {
				return fmt.Errorf("Stok %s habis/kurang! Sisa Stok: %d", product.NamaProduk, product.Stok)
			}

			// Kurangi stok di database
			product.Stok -= item.Kuantitas
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			// Hitung Harga * Kuantitas
			itemSubTotal := product.HargaJual * float64(item.Kuantitas)
			subTotal += itemSubTotal

			// masukan ke body struk (detail)
			details = append(details, models.TransactionDetail{
				ProductID: product.ID,
				HargaSatuan: product.HargaJual,
				Kuantitas: item.Kuantitas,
				SubTotal: itemSubTotal,
			})
		}

		// 3. Logika Keuangan PPN & Pembulatan
		pajak := (store.PajakPersen / 100.0) * subTotal
		rawTotal := subTotal + pajak

		// Pembulatan ke ratusan terdekat
		roundedTotal := math.Round(rawTotal/100) * 100
		pembulatan := roundedTotal - rawTotal

		// Cek Uang pembeli cukup atau kurang
		kembalian := input.NominalBayar - roundedTotal
		if kembalian < 0 {
			return fmt.Errorf("Uang pelanggan kurang Rp %.0f !", math.Abs(kembalian))
		}

		// 4. Generate Nomor Struk / Invoice
		noInvoice := fmt.Sprintf("INV-%s", time.Now().Format("20060102150405"))

		// 5. Simpan header struk
		savedTransaction = models.Transaction{
			StoreID:      uint(storeID.(float64)),
			UserID:       uint(userID.(float64)),
			NoInvoice:    noInvoice,
			SubTotal:     subTotal,
			Pajak:        pajak,
			Pembulatan:   pembulatan,
			TotalHarga:   roundedTotal,
			NominalBayar: input.NominalBayar,
			Kembalian:    kembalian,
			Details:      details, // GORM otomatis masukin detail ini ke tabel sebelahnya!
		}

		// Eksekusi Penyimpanan
		if err := tx.Create(&savedTransaction).Error; err != nil {
			return err
		}

		return nil //sukses disini berarti data commit permanen
	})
	// --- Selesai Database TRX ---

	// Ditengah jalan stok kurang atau uang kurang, reject permintaan
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Kalau Sukses, kirim struk ke layar kasir
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaksi berhasil! 💸 Struk siap dicetak.",
		"invoice": savedTransaction.NoInvoice,
		"tagihan": savedTransaction.TotalHarga,
		"kembali": savedTransaction.Kembalian,
	})
}