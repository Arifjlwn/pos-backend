package controllers

import (
	"net/http"
	"pos-backend/config"
	"pos-backend/models"

	"github.com/gin-gonic/gin"
)

// Struct untuk menangkap data dari Frontend
type ProductInput struct {
	SKU        *string `json:"sku"`
	NamaProduk string  `json:"nama_produk" binding:"required"`
	Kategori   string  `json:"kategori"`
	HargaModal float64 `json:"harga_modal"`
	HargaJual  float64 `json:"harga_jual" binding:"required"`
	Stok       int     `json:"stok"`
}

// Fungsi Tambah Produk
func CreateProduct(c *gin.Context) {
	// 1. Ambil ID Toko dari token JWT yang lolos satpam
	storeID, exists := c.Get("store_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak, ID Toko tidak ditemukan!"})
		return
	}

	// 2. Tangkap JSON dari Frontend
	var input ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Rakit data produknya
	product := models.Product{
		// Convert storeID ke tipe uint (karena JWT nyimpennya float64 secara default)
		StoreID:    uint(storeID.(float64)),
		SKU:        input.SKU,
		NamaProduk: input.NamaProduk,
		Kategori:   input.Kategori,
		HargaModal: input.HargaModal,
		HargaJual:  input.HargaJual,
		Stok:       input.Stok,
	}

	// 4. Simpan ke database
	if err := config.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk. SKU mungkin bentrok."})
		return
	}

	// 5. Beri balasan sukses
	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil ditambahkan! 📦",
		"data":    product,
	})
}

// Fungsi Lihat Daftar Produk
func GetProducts(c *gin.Context) {
	// Tanya Satpam: "Ini yang lagi login ID toko berapa ?"
	storeID, exists := c.Get("store_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses Ditolak !"})
		return
	}

	// Siapkan wadah (Array/Slice) untuk menampung banyak produk
	var products []models.Product

	// Cari DB semua produk yang store id nya cocok
	if err := config.DB.Where("store_id = ?", storeID).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	// Kirim data ke FE
	c.JSON(http.StatusOK, gin.H{
		"message": "Katalog produk berhasil dimuat! 📚",
		"total":   len(products), // Ngitung ada berapa barang
		"data":    products,
	})
}

// Fungsi Ubah Produk (Update)
func UpdateProduct(c *gin.Context) {
	// 1. Cek ID Toko dari Satpam JWT
	storeID, _ := c.Get("store_id")
	role, _ := c.Get("role")

	// Logika RBAC
	if role != "owner" {
		// Status 403 Forbidden (Dilarang Masuk)
		c.JSON(http.StatusForbidden, gin.H{"error": "Hentikan! Cuma Owner yang boleh ubah harga/data barang."})
		return
	}
	
	// 2. Tangkap ID Produk dari ujung URL (Contoh: /api/products/1)
	productID := c.Param("id") 
	var product models.Product
	
	// 3. Cari produknya. Syarat Wajib: ID Produk harus cocok DAN ID Toko harus cocok!
	if err := config.DB.Where("id = ? AND store_id = ?", productID, storeID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan atau bukan milik toko Anda!"})
		return
	}

	// 4. Tangkap data baru dari Frontend
	var input ProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 5. Timpa data lama dengan data baru
	product.SKU = input.SKU
	product.NamaProduk = input.NamaProduk
	product.Kategori = input.Kategori
	product.HargaModal = input.HargaModal
	product.HargaJual = input.HargaJual
	product.Stok = input.Stok

	// 6. Simpan kembali ke database
	config.DB.Save(&product)

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil diubah! ✏️", 
		"data": product,
	})
}

// Fungsi Hapus Produk (Delete)
func DeleteProduct(c *gin.Context) {
	storeID, _ := c.Get("store_id")
	role, _ := c.Get("role")

	// Logika RBAC
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Waduh, Kasir dilarang hapus barang dari sistem!"})
		return
	}

	productID := c.Param("id")
	var product models.Product
	
	// Pastikan produk yang mau dihapus itu beneran ada dan milik dia
	if err := config.DB.Where("id = ? AND store_id = ?", productID, storeID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan atau bukan milik toko Anda!"})
		return
	}

	// Hapus dari muka bumi (database)
	config.DB.Delete(&product)
	
	c.JSON(http.StatusOK, gin.H{"message": "Barang berhasil dihapus dari gudang! 🗑️"})
}