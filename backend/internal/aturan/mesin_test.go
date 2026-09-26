package aturan

import (
	"database/sql"
	"testing"

	"sivilize-traffic-routing-demo/backend/internal/model"
)

// Helper untuk membuat sql.NullString yang valid
func strNull(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// 1. ID + mobile cocok dengan rule ID + mobile
func TestIDDanMobileCocok(t *testing.T) {
	rule := model.Aturan{
		ID:        1,
		Nama:      "Rule ID Mobile",
		Negara:    strNull("ID"),
		Perangkat: strNull("mobile"),
		Status:    true,
		TujuanURL: "/demo/halaman-a",
	}

	konteks := KonteksPengunjung{
		Negara:    "ID",
		Perangkat: "mobile",
		Peramban:  "Chrome",
	}

	if !CocokkanAturan(rule, konteks) {
		t.Errorf("Harusnya cocok untuk ID + mobile")
	}
}

// 2. ID + desktop tidak cocok dengan rule ID + mobile
func TestIDDanDesktopTidakCocokDenganRuleMobile(t *testing.T) {
	rule := model.Aturan{
		ID:        1,
		Nama:      "Rule ID Mobile",
		Negara:    strNull("ID"),
		Perangkat: strNull("mobile"),
		Status:    true,
	}

	konteks := KonteksPengunjung{
		Negara:    "ID",
		Perangkat: "desktop",
	}

	if CocokkanAturan(rule, konteks) {
		t.Errorf("Harusnya tidak cocok untuk desktop dengan rule mobile")
	}
}

// 3. Rule negara NULL dapat cocok dengan negara apa pun
func TestRuleNegaraNullCocokSemuaNegara(t *testing.T) {
	rule := model.Aturan{
		ID:        1,
		Nama:      "Rule Bebas Negara",
		Negara:    sql.NullString{Valid: false},
		Perangkat: strNull("desktop"),
		Status:    true,
	}

	daftarNegara := []string{"ID", "US", "SG", "JP", ""}
	for _, neg := range daftarNegara {
		k := KonteksPengunjung{Negara: neg, Perangkat: "desktop"}
		if !CocokkanAturan(rule, k) {
			t.Errorf("Rule negara NULL harusnya cocok untuk negara '%s'", neg)
		}
	}
}

// 4. Rule perangkat NULL dapat cocok dengan perangkat apa pun
func TestRulePerangkatNullCocokSemuaPerangkat(t *testing.T) {
	rule := model.Aturan{
		ID:        1,
		Nama:      "Rule Bebas Perangkat",
		Negara:    strNull("ID"),
		Perangkat: sql.NullString{Valid: false},
		Status:    true,
	}

	daftarPerangkat := []string{"mobile", "tablet", "desktop", "unknown"}
	for _, p := range daftarPerangkat {
		k := KonteksPengunjung{Negara: "ID", Perangkat: p}
		if !CocokkanAturan(rule, k) {
			t.Errorf("Rule perangkat NULL harusnya cocok untuk perangkat '%s'", p)
		}
	}
}

// 5. Rule browser NULL dapat cocok dengan browser apa pun
func TestRuleBrowserNullCocokSemuaBrowser(t *testing.T) {
	rule := model.Aturan{
		ID:       1,
		Nama:     "Rule Bebas Browser",
		Negara:   strNull("ID"),
		Peramban: sql.NullString{Valid: false},
		Status:   true,
	}

	daftarPeramban := []string{"Chrome", "Firefox", "Safari", "Edge", "Lainnya", "Tidak Diketahui"}
	for _, b := range daftarPeramban {
		k := KonteksPengunjung{Negara: "ID", Peramban: b}
		if !CocokkanAturan(rule, k) {
			t.Errorf("Rule peramban NULL harusnya cocok untuk browser '%s'", b)
		}
	}
}

// 6. Semua kondisi terisi harus cocok semuanya
func TestSemuaKondisiTerisiHarusCocokSemuanya(t *testing.T) {
	rule := model.Aturan{
		ID:           1,
		Nama:         "Rule Ketat",
		Negara:       strNull("ID"),
		Perangkat:    strNull("mobile"),
		Peramban:     strNull("Chrome"),
		AgenPengguna: strNull("Pixel"),
		AsalRujukan:  strNull("google.com"),
		Status:       true,
	}

	// Kasus cocok
	kCocok := KonteksPengunjung{
		Negara:       "ID",
		Perangkat:    "mobile",
		Peramban:     "Chrome",
		AgenPengguna: "Mozilla/5.0 Google Pixel 7",
		AsalRujukan:  "https://www.google.com/search",
	}
	if !CocokkanAturan(rule, kCocok) {
		t.Errorf("Semua kondisi sama, harusnya cocok")
	}

	// Salah satu kondisi berbeda (misal peramban Firefox)
	kTidakCocok := kCocok
	kTidakCocok.Peramban = "Firefox"
	if CocokkanAturan(rule, kTidakCocok) {
		t.Errorf("Harusnya tidak cocok jika salah satu kondisi tidak sesuai")
	}
}

// 7. Rule inactive tidak digunakan
func TestRuleInactiveTidakDigunakan(t *testing.T) {
	ruleNonaktif := model.Aturan{
		ID:        1,
		Nama:      "Rule Nonaktif",
		Negara:    strNull("ID"),
		Status:    false,
		Prioritas: 1,
		TujuanURL: "/demo/halaman-a",
	}

	k := KonteksPengunjung{Negara: "ID"}
	if CocokkanAturan(ruleNonaktif, k) {
		t.Errorf("Aturan nonaktif tidak boleh cocok")
	}

	hasil := EvaluasiKonteks([]model.Aturan{ruleNonaktif}, k, "/demo/cadangan")
	if hasil.Cocok || hasil.Hasil != "fallback" {
		t.Errorf("Aturan nonaktif harusnya mengarah ke fallback")
	}
}

// 8. Tidak ada rule → fallback
func TestTidakAdaRuleMenghasilkanFallback(t *testing.T) {
	k := KonteksPengunjung{Negara: "US"}
	hasil := EvaluasiKonteks([]model.Aturan{}, k, "/demo/cadangan")

	if hasil.Cocok || hasil.Hasil != "fallback" || hasil.URLTujuan != "/demo/cadangan" {
		t.Errorf("Harusnya fallback saat tidak ada aturan")
	}
}

// 9. Dua rule cocok → priority terkecil menang
func TestDuaRuleCocokPriorityTerkecilMenang(t *testing.T) {
	rule1 := model.Aturan{
		ID:        10,
		Nama:      "Rule Prioritas 3",
		Negara:    strNull("ID"),
		Prioritas: 3,
		Status:    true,
		TujuanURL: "/demo/halaman-p3",
	}
	rule2 := model.Aturan{
		ID:        20,
		Nama:      "Rule Prioritas 1",
		Negara:    strNull("ID"),
		Prioritas: 1,
		Status:    true,
		TujuanURL: "/demo/halaman-p1",
	}

	k := KonteksPengunjung{Negara: "ID"}
	hasil := EvaluasiKonteks([]model.Aturan{rule1, rule2}, k, "/demo/cadangan")

	if !hasil.Cocok || *hasil.AturanID != 20 || hasil.URLTujuan != "/demo/halaman-p1" {
		t.Errorf("Harusnya rule prioritas 1 menang, didapat ID: %v", hasil.AturanID)
	}
}

// 10. Priority 1 menang atas priority 5
func TestPriority1MenangAtasPriority5(t *testing.T) {
	rP1 := model.Aturan{
		ID:        1,
		Nama:      "Prioritas 1",
		Negara:    strNull("ID"),
		Prioritas: 1,
		Status:    true,
		TujuanURL: "/demo/p1",
	}
	rP5 := model.Aturan{
		ID:        5,
		Nama:      "Prioritas 5",
		Negara:    strNull("ID"),
		Prioritas: 5,
		Status:    true,
		TujuanURL: "/demo/p5",
	}

	k := KonteksPengunjung{Negara: "ID"}
	hasil := EvaluasiKonteks([]model.Aturan{rP5, rP1}, k, "/demo/cadangan")

	if *hasil.AturanID != 1 {
		t.Errorf("Priority 1 harus menang atas Priority 5, didapat ID: %v", hasil.AturanID)
	}
}

// 11. User-Agent mobile terdeteksi mobile
func TestDeteksiPerangkatMobile(t *testing.T) {
	uaiPhone := "Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"
	uaAndroidMobile := "Mozilla/5.0 (Linux; Android 13; SM-S908B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36"

	if DeteksiPerangkat(uaiPhone) != "mobile" {
		t.Errorf("iPhone harus terdeteksi mobile")
	}
	if DeteksiPerangkat(uaAndroidMobile) != "mobile" {
		t.Errorf("Android Mobile harus terdeteksi mobile")
	}
}

// 12. User-Agent tablet terdeteksi tablet
func TestDeteksiPerangkatTablet(t *testing.T) {
	uaiPad := "Mozilla/5.0 (iPad; CPU OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Mobile/15E148 Safari/604.1"
	uaAndroidTablet := "Mozilla/5.0 (Linux; Android 12; SM-T870) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"

	if DeteksiPerangkat(uaiPad) != "tablet" {
		t.Errorf("iPad harus terdeteksi tablet")
	}
	if DeteksiPerangkat(uaAndroidTablet) != "tablet" {
		t.Errorf("Android Tablet (tanpa token Mobile) harus terdeteksi tablet")
	}
}

// 13. User-Agent desktop terdeteksi desktop
func TestDeteksiPerangkatDesktop(t *testing.T) {
	uaWindows := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
	uaMac := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	uaLinux := "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/114.0"

	if DeteksiPerangkat(uaWindows) != "desktop" {
		t.Errorf("Windows harus terdeteksi desktop")
	}
	if DeteksiPerangkat(uaMac) != "desktop" {
		t.Errorf("Mac harus terdeteksi desktop")
	}
	if DeteksiPerangkat(uaLinux) != "desktop" {
		t.Errorf("Linux harus terdeteksi desktop")
	}
}

// 14. Edge tidak salah dikategorikan sebagai Chrome
func TestEdgeTidakSalahSebagaiChrome(t *testing.T) {
	uaEdge := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36 Edg/114.0.1823.58"
	peramban := DeteksiPeramban(uaEdge)

	if peramban != "Edge" {
		t.Errorf("User-Agent Edge harus terdeteksi sebagai 'Edge', bukan '%s'", peramban)
	}
}

// 15. Firefox terdeteksi dengan benar
func TestDeteksiFirefox(t *testing.T) {
	uaFirefox := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0"
	if DeteksiPeramban(uaFirefox) != "Firefox" {
		t.Errorf("Firefox harus terdeteksi sebagai 'Firefox'")
	}
}

// 16. Safari terdeteksi dengan benar
func TestDeteksiSafari(t *testing.T) {
	uaSafari := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15"
	peramban := DeteksiPeramban(uaSafari)

	if peramban != "Safari" {
		t.Errorf("Safari harus terdeteksi sebagai 'Safari', bukan '%s'", peramban)
	}
}

// 17. Browser tidak dikenal menghasilkan Tidak Diketahui
func TestBrowserTidakDikenal(t *testing.T) {
	if DeteksiPeramban("") != "Tidak Diketahui" {
		t.Errorf("User-Agent kosong harus 'Tidak Diketahui'")
	}
	if DeteksiPeramban("   ") != "Tidak Diketahui" {
		t.Errorf("User-Agent spasi harus 'Tidak Diketahui'")
	}
	if DeteksiPeramban("PerangkatKustom/1.0") != "Tidak Diketahui" {
		t.Errorf("User-Agent asing tanpa petunjuk harus 'Tidak Diketahui'")
	}
}

// 18. Normalisasi negara berhasil
func TestNormalisasiNegara(t *testing.T) {
	if NormalisasiNegara("id") != "ID" {
		t.Errorf("'id' harus dinormalisasi menjadi 'ID'")
	}
	if NormalisasiNegara("Id") != "ID" {
		t.Errorf("'Id' harus dinormalisasi menjadi 'ID'")
	}
	if NormalisasiNegara("  id  ") != "ID" {
		t.Errorf("'  id  ' harus dinormalisasi menjadi 'ID'")
	}
}

// 19. Normalisasi perangkat berhasil
func TestNormalisasiPerangkat(t *testing.T) {
	if NormalisasiPerangkat("Mobile") != "mobile" {
		t.Errorf("'Mobile' harus dinormalisasi menjadi 'mobile'")
	}
	if NormalisasiPerangkat("MOBILE") != "mobile" {
		t.Errorf("'MOBILE' harus dinormalisasi menjadi 'mobile'")
	}
	if NormalisasiPerangkat("  Desktop  ") != "desktop" {
		t.Errorf("'  Desktop  ' harus dinormalisasi menjadi 'desktop'")
	}
}

// 20. Normalisasi browser berhasil
func TestNormalisasiBrowser(t *testing.T) {
	if NormalisasiPeramban("chrome") != "Chrome" {
		t.Errorf("'chrome' harus 'Chrome'")
	}
	if NormalisasiPeramban("FIREFOX") != "Firefox" {
		t.Errorf("'FIREFOX' harus 'Firefox'")
	}
	if NormalisasiPeramban("safari") != "Safari" {
		t.Errorf("'safari' harus 'Safari'")
	}
	if NormalisasiPeramban("edge") != "Edge" {
		t.Errorf("'edge' harus 'Edge'")
	}
	if NormalisasiPeramban("unknown") != "Tidak Diketahui" {
		t.Errorf("'unknown' harus 'Tidak Diketahui'")
	}
}

// 21. Rule dengan referrer cocok
func TestRuleDenganReferrerCocok(t *testing.T) {
	rule := model.Aturan{
		ID:          1,
		Nama:        "Rule Dari Google",
		AsalRujukan: strNull("google.com"),
		Status:      true,
	}

	kCocok := KonteksPengunjung{AsalRujukan: "https://www.google.com/search?q=demo"}
	if !CocokkanAturan(rule, kCocok) {
		t.Errorf("Harusnya cocok dengan referrer google.com")
	}

	kBeda := KonteksPengunjung{AsalRujukan: "https://bing.com"}
	if CocokkanAturan(rule, kBeda) {
		t.Errorf("Harusnya tidak cocok dengan referrer bing.com")
	}
}

// 22. Rule dengan referrer NULL mengabaikan referrer
func TestRuleReferrerNullMengabaikanReferrer(t *testing.T) {
	rule := model.Aturan{
		ID:          1,
		Nama:        "Rule Bebas Referrer",
		AsalRujukan: sql.NullString{Valid: false},
		Status:      true,
	}

	daftarRef := []string{"", "https://google.com", "https://facebook.com"}
	for _, ref := range daftarRef {
		k := KonteksPengunjung{AsalRujukan: ref}
		if !CocokkanAturan(rule, k) {
			t.Errorf("Rule referrer NULL harusnya mengabaikan referrer '%s'", ref)
		}
	}
}

// 23. Rule dengan User-Agent cocok
func TestRuleDenganUserAgentCocok(t *testing.T) {
	rule := model.Aturan{
		ID:           1,
		Nama:         "Rule Android",
		AgenPengguna: strNull("Android"),
		Status:       true,
	}

	kCocok := KonteksPengunjung{AgenPengguna: "Mozilla/5.0 (Linux; Android 13; Pixel 7)"}
	if !CocokkanAturan(rule, kCocok) {
		t.Errorf("Harusnya cocok dengan substring Android")
	}

	kBeda := KonteksPengunjung{AgenPengguna: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_5)"}
	if CocokkanAturan(rule, kBeda) {
		t.Errorf("Harusnya tidak cocok dengan iPhone")
	}
}

// 24. Rule dengan User-Agent NULL mengabaikan User-Agent
func TestRuleUserAgentNullMengabaikanUserAgent(t *testing.T) {
	rule := model.Aturan{
		ID:           1,
		Nama:         "Rule Bebas UA",
		AgenPengguna: sql.NullString{Valid: false},
		Status:       true,
	}

	daftarUA := []string{"", "curl/7.68.0", "Mozilla/5.0"}
	for _, ua := range daftarUA {
		k := KonteksPengunjung{AgenPengguna: ua}
		if !CocokkanAturan(rule, k) {
			t.Errorf("Rule UA NULL harusnya mengabaikan UA '%s'", ua)
		}
	}
}

// 25. Destination tidak ditemukan ditangani dengan aman (fallback)
func TestTujuanTidakDitemukanDitanganiAman(t *testing.T) {
	ruleTanpaURL := model.Aturan{
		ID:        1,
		Nama:      "Rule Tujuan Hilang",
		Negara:    strNull("ID"),
		Prioritas: 1,
		Status:    true,
		TujuanURL: "", // Tujuan sudah dihapus / tidak valid
	}

	k := KonteksPengunjung{Negara: "ID"}
	hasil := EvaluasiKonteks([]model.Aturan{ruleTanpaURL}, k, "/demo/cadangan")

	if hasil.Cocok || hasil.Hasil != "fallback" || hasil.URLTujuan != "/demo/cadangan" {
		t.Errorf("Tujuan kosong harus ditangani dengan fallback secara aman: %+v", hasil)
	}
}

// 26. Beberapa rule dengan priority sama menghasilkan keputusan deterministic (ID terkecil)
func TestPrioritySamaDeterministik(t *testing.T) {
	r1 := model.Aturan{
		ID:        99,
		Nama:      "Rule ID 99",
		Negara:    strNull("ID"),
		Prioritas: 5,
		Status:    true,
		TujuanURL: "/demo/99",
	}
	r2 := model.Aturan{
		ID:        12,
		Nama:      "Rule ID 12",
		Negara:    strNull("ID"),
		Prioritas: 5,
		Status:    true,
		TujuanURL: "/demo/12",
	}
	r3 := model.Aturan{
		ID:        45,
		Nama:      "Rule ID 45",
		Negara:    strNull("ID"),
		Prioritas: 5,
		Status:    true,
		TujuanURL: "/demo/45",
	}

	k := KonteksPengunjung{Negara: "ID"}

	// Jalankan berulang-ulang untuk memastikan determinisme
	for i := 0; i < 5; i++ {
		hasil := EvaluasiKonteks([]model.Aturan{r1, r2, r3}, k, "/demo/cadangan")
		if *hasil.AturanID != 12 {
			t.Fatalf("Harus deterministik memilih ID terkecil (12), didapat: %d", *hasil.AturanID)
		}
	}
}
