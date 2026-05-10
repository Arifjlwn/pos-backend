package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(c *gin.Context) {
	// Cek apakah ada karcis di kantong (Header Auth)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak ! Anda belum login."})
		c.Abort() //Langsung Usir, Jangan kasih masuk
		return
	}

	// Format Karcis biasanya diawali "Bearer ", jadi potong sisa kode token aja
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// Cek Keaslian Karcis pakai kunci rahasia saat login
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metode token tidak valid")
		}
		return []byte("KUNCI_RAHASIA_SUPER_KUAT_123"), nil // harus Sama Persis dengan di authController
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau sudah hangus !"})
		c.Abort()
		return
	}

	// Kalau karcis asli, bongkar isi nya (Ambil ID User, ID TOKO, dan Role)
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Simpan data ini kedalam "Context" agar bisa dibaca oleh controller nanti
		c.Set("user_id", claims["user_id"])
		c.Set("store_id", claims["store_id"])
		c.Set("role", claims["role"])

		c.Next() // Silahkan Masuk !
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Gagal membaca data dari token"})
		c.Abort()
		return
	}
}