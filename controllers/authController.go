package controllers

import(
	"net/http"
	"pos-backend/config"
	"pos-backend/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Ini adalah struktur keranjang untuk menangkan JSON yg dikirim frontend
type RegisterInput struct {
	NamaToko     string `json:"nama_toko" binding:"required"`
	Alamat       string `json:"alamat"`
	Telepon      string `json:"telepon"`
	BusinessType string `json:"business_type" binding:"required"` // kelontong, cafe, laundry
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
}
// --- Register ---
func Register(c *gin.Context) {
	var input RegisterInput

	// Tangkap JSON dan Validasi (Harus sesuai dengan struct RegisterInput)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	// Acak Password Biar Aman
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password),bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal Mengenkripsi password"})
		return
	}

	// Simpan data toko (Store) ke database
	store := models.Store{
		NamaToko:     input.NamaToko,
		Alamat:       input.Alamat,
		Telepon:      input.Telepon,
		BusinessType: input.BusinessType,
	}
	config.DB.Create(&store)

	// Simpan data User dan kaitkan dengan ID Toko yang baru terbuat
	user := models.User{
		StoreID:  store.ID,
		Name:     input.Name,
		Email:    &input.Email, 
		Password: string(hashedPassword),
		Role:     "owner",
	}
	config.DB.Create(&user)

	// Kembalikan status sukses ke FE
	c.JSON(http.StatusCreated, gin.H{
		"message": "Registrasi Toko & Akun berhasil! 🚀",
		"store":   store.NamaToko,
		"owner":   user.Name,
		"type":    store.BusinessType,
	})
}

// --- Login ---
// Strung Penangkap data login
type LoginInput struct {
	Identifier string `json:"identifier" binding:"required"` // Bisa isi Email / NIK
	Password   string `json:"password" binding:"required"`
}

// Fungsi
func Login(c *gin.Context) {
	var input LoginInput

	// Tangkap inputan email & password
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Deteksi Login
	if strings.Contains(input.Identifier, "@") {
		// Jika input ada @ berarti cari berdasarkan email
		if err := config.DB.Where("email = ?", input.Identifier).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Email tidak terdaftar !"})
			return
		}
	} else {
		// Jika Tidak ada @ cari berdasarkan NIK
		if err := config.DB.Where("nik = ?", input.Identifier).First(&user).Error;err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "NIK tidak ditemukan !"})
			return
		}
	}

	// Cocokkan Password yang diinput dengan password acak di database
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah !"})
		return
	}

	// PEMBUATAN TIKET (JWT)
	// Kita simpan ID User dan ID Toko di dalam tiket ini
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"store_id": user.StoreID, // Ini penting banget buat nyari barang toko dia aja!
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 72).Unix(), // Karcis hangus dalam 3 hari (72 Jam)
	})

	// Tanda tangani tiket pakai Kunci Rahasia
	// (Nanti di versi asli kita taruh kunci ini di file .env biar lebih aman)
	tokenString, err := token.SignedString([]byte("KUNCI_RAHASIA_SUPER_KUAT_123"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencetak tiket masuk"})
		return
	}

	// Kembalikan Tiket (Token) ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Login Sukses! Selamat datang Bos.",
		"token":   tokenString,
		"role": user.Role, //Beritahu FE ini owner atau kasir
	})

}
