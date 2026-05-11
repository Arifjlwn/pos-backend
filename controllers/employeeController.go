package controllers

import (
	"fmt"
	"net/http"
	"pos-backend/config"
	"pos-backend/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Struct penangkap data dari Bos
type EmployeeInput struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// Fungsi Menambahkan Karyawan
func CreateEmployee(c *gin.Context) {
	// 1. Cek siapa yg lagi akses (wajib Owner)
	storeID, _ := c.Get("store_id")
	role, _ := c.Get("role")

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Owner yang bisa menambahkan karyawan baru !"})
		return
	}

	// 2. Tangkap inputan nama dan password sementara dari bos
	var input EmployeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. LOGIKA GENERATE NIK OTOMATIS
	// Ambil tahun saat ini
	currentYear := time.Now().Format("2006")

	var lastEmployee models.User
	var newNIK string

	// cari karyawan kasir terakhir, urutkan berdasarkan NIK dari kecil ke besa ( DESC )
	err := config.DB.Where("store_id = ? AND role = ? AND nik LIKE ?", storeID, "kasir", currentYear+"%").
	Order("nik desc").
	First(&lastEmployee).Error

	if err != nil {
		// Kalau error berarti dia karyawan pertama
		newNIK = currentYear + "0001"
	} else {
		// Kalau ketemu, ambil NIK terakhir nya
		lastNIK := *lastEmployee.NIK

		// Potong 4 digit terakhir
		lastSequenceStr := lastNIK[4:]

		// Ubah jadi angka (integer)
		lastSequence, _ := strconv.Atoi(lastSequenceStr)

		// Tambah 1 untuk karyawan baru
		newSequence := lastSequence + 1

		// Gabungkan kembali
		newNIK = fmt.Sprintf("%s%04d", currentYear, newSequence)
	}

	// 4. Acak password sementara kasir
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gagal mengenkripsi password"})
		return
	}

	// 5. Simpan ke database
	employee := models.User{
		StoreID:  uint(storeID.(float64)),
		Name:     input.Name,
		NIK:      &newNIK, // Pakai & karena di model tipe datanya pointer (*string)
		Password: string(hashedPassword),
		Role:     "kasir",
	}

	config.DB.Create(&employee)

	// 6. Kembalikan data ke frontend
	c.JSON(http.StatusCreated, gin.H{
		"message": "Karyawan baru berhasil didaftarkan! 🤝",
		"data": gin.H{
			"nama": employee.Name,
			"nik":  newNIK, // Tampilkan NIK agar Bos bisa ngasih tau ke karyawannya
			"role": employee.Role,
		},
	})
}

// Fungsi melihat daftar karyawan (Khusus Owner)
// Fungsi Lihat Daftar Karyawan (Khusus Bos)
func GetEmployees(c *gin.Context) {
	// 1. Ambil ID Toko dan Role dari Satpam JWT
	storeID, _ := c.Get("store_id")
	role, _ := c.Get("role")

	// 2. Gembok pintu! Cuma Owner yang boleh lihat data rahasia ini
	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses ditolak! Hanya Bos yang boleh melihat daftar karyawan."})
		return
	}

	// 3. Siapkan wadah untuk daftar karyawan
	var employees []models.User

	// 4. Cari semua kasir yang bekerja di toko tersebut
	// (Note: Password tidak akan ikut terkirim ke Frontend karena di models/user.go sudah kita beri json:"-")
	if err := config.DB.Where("store_id = ? AND role = ?", storeID, "kasir").Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data karyawan"})
		return
	}

	// 5. Kirim datanya ke Frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Daftar karyawan berhasil dimuat! 👥",
		"total":   len(employees),
		"data":    employees,
	})
}