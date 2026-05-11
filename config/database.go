package config

import (
	"log"
	"pos-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Variabel Global untuk menampung Koneksi DB
var DB *gorm.DB

func ConnectDatabase() {
	// Format: username:password@tcp(host:port)/nama_database?opsi_tambahan
	// Sesuaikan jika MYSQL pakai password
	dsn := "root:@tcp(127.0.0.1:3306)/pos_saas?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal Menyambung ke database ! Error: ", err)
	}

	log.Println("✅ Berhasil terhubung ke Database pos_saas!")

	// Auto Migrate
	err = database.AutoMigrate(
		&models.Store{},
		&models.User{},
		&models.Product{},
		&models.Transaction{},
		&models.TransactionDetail{},
	)
	if err != nil {
		log.Fatal("Gagal Melakukan Migrasi Database ! Error: ", err)
	}
	log.Println("✅ Tabel database berhasil di-generate!")

	DB = database
}