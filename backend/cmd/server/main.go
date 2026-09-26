package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"sivilize-traffic-routing-demo/backend/internal/basisdata"
	"sivilize-traffic-routing-demo/backend/internal/konfigurasi"
	"sivilize-traffic-routing-demo/backend/internal/middleware"
	"sivilize-traffic-routing-demo/backend/internal/penangan"
)

func main() {
	konf := konfigurasi.MuatKonfigurasi()

	// Buka koneksi database — program berhenti jika database tidak tersedia
	db := basisdata.BukaKoneksi(konf.UrlDatabase)
	defer db.Close()

	// Tangani sinyal shutdown agar koneksi database ditutup dengan benar
	selesai := make(chan os.Signal, 1)
	signal.Notify(selesai, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-selesai
		log.Println("Mematikan server...")
		db.Close()
		os.Exit(0)
	}()

	mesin := gin.Default()
	mesin.Use(middleware.AturCORS(konf.UrlFrontend))

	api := mesin.Group("/api")
	{
		// Pemeriksaan kesehatan
		api.GET("/kesehatan", penangan.PeriksaKesehatan(db))

		// CRUD Tujuan
		api.GET("/tujuan", penangan.DaftarTujuan(db))
		api.GET("/tujuan/:id", penangan.SatuTujuan(db))
		api.POST("/tujuan", penangan.BuatTujuan(db))
		api.PUT("/tujuan/:id", penangan.UbahTujuan(db))
		api.DELETE("/tujuan/:id", penangan.HapusTujuan(db))

		// CRUD Aturan
		api.GET("/aturan", penangan.DaftarAturan(db))
		api.GET("/aturan/:id", penangan.SatuAturan(db))
		api.POST("/aturan", penangan.BuatAturan(db))
		api.PUT("/aturan/:id", penangan.UbahAturan(db))
		api.PATCH("/aturan/:id/status", penangan.UbahStatusAturan(db))
		api.DELETE("/aturan/:id", penangan.HapusAturan(db))

		// Pemeriksaan Trafik & Traffic Logs (Batch 5)
		api.POST("/periksa", penangan.PeriksaTrafik(db, konf.UrlFallback))
		api.GET("/log-trafik", penangan.DaftarLogTrafik(db))
		api.GET("/log-trafik/:id", penangan.SatuLogTrafik(db))
	}

	log.Printf("Server Sivilize Traffic Routing Demo berjalan pada port :%s", konf.Port)
	if err := mesin.Run(":" + konf.Port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
