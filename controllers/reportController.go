package controllers

import (
	"net/http"
	"pos-backend/config"
	"pos-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetDashboardReport(c *gin.Context) {
	storeID, _ := c.Get("store_id")
	role, _ := c.Get("role")

	// 1. Gembok! Cuma Bos yang boleh lihat laporan keuangan
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak! Laporan keuangan cuma untuk Owner."})
		return
	}

	// 2. Tentukan rentang waktu "Hari Ini" (00:00:00 sampai 23:59:59)
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// --- LOGIKA HITUNG OMZET & JUMLAH TRANSAKSI ---
	var report struct {
		OmzetHariIni        float64 `json:"OmzetHariIni"`
		JumlahTransaksi      int64   `json:"jumlah_transaksi"`
		TotalProdukTerjual   float64 `json:"total_produk_terjual"`
	}

	// Hitung Omzet & Jumlah Transaksi hari ini
	config.DB.Model(&models.Transaction{}).
		Where("store_id = ? AND created_at BETWEEN ? AND ?", storeID, startOfDay, endOfDay).
		Select("SUM(total_harga) as omzet_hari_ini, COUNT(id) as jumlah_transaksi").
		Scan(&report)

	// --- LOGIKA HITUNG TOTAL PRODUK TERJUAL ---
	// Kita join ke tabel detail transaksi
	config.DB.Table("transaction_details").
		Joins("JOIN transactions ON transactions.id = transaction_details.transaction_id").
		Where("transactions.store_id = ? AND transactions.created_at BETWEEN ? AND ?", storeID, startOfDay, endOfDay).
		Select("SUM(transaction_details.kuantitas)").
		Row().Scan(&report.TotalProdukTerjual)

	// --- LOGIKA STOK MENIPIS (ALERT) ---
	var lowStockProducts []models.Product
	// Kita ambil produk yang stoknya < 10 (Mas bisa ganti angka ini sesuai kebutuhan)
	config.DB.Where("store_id = ? AND stok < ?", storeID, 10).
		Find(&lowStockProducts)

	// 3. Kirim hasil laporan ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Data laporan berhasil ditarik! 📊",
		"data": gin.H{
			"summary": report,
			"low_stock": lowStockProducts,
		},
	})
}