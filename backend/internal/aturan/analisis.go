package aturan

import (
	"net"
	"net/http"
	"strings"
)

// KonteksPengunjung menyimpan atribut hasil analisis dari HTTP request atau simulasi.
type KonteksPengunjung struct {
	Negara       string `json:"negara"`
	Perangkat    string `json:"perangkat"`
	Peramban     string `json:"peramban"`
	AgenPengguna string `json:"agen_pengguna"`
	AsalRujukan  string `json:"asal_rujukan"`
	IP           string `json:"ip"`
}

// NormalisasiNegara memastikan kode negara dalam format huruf kapital tanpa spasi (misal: "id" -> "ID").
func NormalisasiNegara(n string) string {
	return strings.ToUpper(strings.TrimSpace(n))
}

// NormalisasiPerangkat memastikan kategori perangkat dalam huruf kecil tanpa spasi (misal: "Mobile" -> "mobile").
func NormalisasiPerangkat(p string) string {
	bersih := strings.ToLower(strings.TrimSpace(p))
	switch bersih {
	case "mobile", "tablet", "desktop":
		return bersih
	case "":
		return "unknown"
	default:
		return bersih
	}
}

// NormalisasiPeramban mengubah nama peramban menjadi bentuk kanonis.
func NormalisasiPeramban(b string) string {
	bersih := strings.TrimSpace(b)
	switch strings.ToLower(bersih) {
	case "chrome":
		return "Chrome"
	case "firefox":
		return "Firefox"
	case "safari":
		return "Safari"
	case "edge":
		return "Edge"
	case "lainnya":
		return "Lainnya"
	case "tidak diketahui", "unknown", "":
		return "Tidak Diketahui"
	default:
		return bersih
	}
}

// DeteksiPerangkat menganalisis string User-Agent untuk mengklasifikasikan jenis perangkat.
func DeteksiPerangkat(ua string) string {
	uaKecil := strings.ToLower(ua)
	if strings.TrimSpace(uaKecil) == "" {
		return "unknown"
	}

	// 1. Deteksi Tablet
	// iPad, PlayBook, Kindle Silk, atau perangkat Android yang tidak memiliki token 'Mobile'
	if strings.Contains(uaKecil, "ipad") ||
		strings.Contains(uaKecil, "playbook") ||
		strings.Contains(uaKecil, "silk") ||
		strings.Contains(uaKecil, "tablet") ||
		(strings.Contains(uaKecil, "android") && !strings.Contains(uaKecil, "mobile")) {
		return "tablet"
	}

	// 2. Deteksi Mobile / Smartphone
	if strings.Contains(uaKecil, "mobile") ||
		strings.Contains(uaKecil, "iphone") ||
		strings.Contains(uaKecil, "ipod") ||
		strings.Contains(uaKecil, "blackberry") ||
		strings.Contains(uaKecil, "bb10") ||
		strings.Contains(uaKecil, "iemobile") ||
		strings.Contains(uaKecil, "opera mini") ||
		strings.Contains(uaKecil, "opera mobi") ||
		strings.Contains(uaKecil, "webos") {
		return "mobile"
	}

	// 3. Deteksi Desktop / Laptop
	if strings.Contains(uaKecil, "windows nt") ||
		strings.Contains(uaKecil, "macintosh") ||
		strings.Contains(uaKecil, "mac os x") ||
		strings.Contains(uaKecil, "x11") ||
		strings.Contains(uaKecil, "cros") ||
		(strings.Contains(uaKecil, "linux") && !strings.Contains(uaKecil, "android")) {
		return "desktop"
	}

	return "unknown"
}

// DeteksiPeramban menganalisis string User-Agent untuk mengidentifikasi peramban.
// Catatan: Edge diprioritaskan sebelum Chrome karena Edge menyertakan token 'Chrome'.
// Chrome diprioritaskan sebelum Safari karena Chrome menyertakan token 'Safari'.
func DeteksiPeramban(ua string) string {
	if strings.TrimSpace(ua) == "" {
		return "Tidak Diketahui"
	}

	uaKecil := strings.ToLower(ua)

	// 1. Edge (Edg/, Edge/, EdgA/, EdgiOS/)
	if strings.Contains(uaKecil, "edg/") ||
		strings.Contains(uaKecil, "edge/") ||
		strings.Contains(uaKecil, "edga/") ||
		strings.Contains(uaKecil, "edgios/") {
		return "Edge"
	}

	// 2. Firefox (Firefox/, FxiOS/)
	if strings.Contains(uaKecil, "firefox/") ||
		strings.Contains(uaKecil, "fxios/") {
		return "Firefox"
	}

	// 3. Chrome (Chrome/, CriOS/) — dipanggil setelah Edge
	if strings.Contains(uaKecil, "chrome/") ||
		strings.Contains(uaKecil, "crios/") {
		return "Chrome"
	}

	// 4. Safari (mengandung Safari dan Version, atau Mobile Safari tanpa Chrome/Edge)
	if strings.Contains(uaKecil, "safari") &&
		(strings.Contains(uaKecil, "version/") || strings.Contains(uaKecil, "mobile/")) {
		return "Safari"
	}

	// 5. Peramban Lainnya jika ada User-Agent namun bukan salah satu dari 4 peramban utama
	if strings.Contains(uaKecil, "opera") ||
		strings.Contains(uaKecil, "opr/") ||
		strings.Contains(uaKecil, "brave") ||
		strings.Contains(uaKecil, "vivaldi") ||
		strings.Contains(uaKecil, "curl") ||
		strings.Contains(uaKecil, "postman") ||
		strings.Contains(uaKecil, "bot") {
		return "Lainnya"
	}

	return "Tidak Diketahui"
}

// AnalisisPermintaan membuat KonteksPengunjung dari parameter yang diberikan.
// Nilai asli AgenPengguna tetap disimpan secara utuh tanpa modifikasi.
func AnalisisPermintaan(userAgent, asalRujukan, ip, negaraInput string) KonteksPengunjung {
	perangkat := DeteksiPerangkat(userAgent)
	peramban := DeteksiPeramban(userAgent)

	return KonteksPengunjung{
		Negara:       NormalisasiNegara(negaraInput),
		Perangkat:    NormalisasiPerangkat(perangkat),
		Peramban:     NormalisasiPeramban(peramban),
		AgenPengguna: userAgent,
		AsalRujukan:  strings.TrimSpace(asalRujukan),
		IP:           strings.TrimSpace(ip),
	}
}

// EkstrakDariHTTPRequest menganalisis objek http.Request secara langsung.
func EkstrakDariHTTPRequest(r *http.Request, negaraInput string) KonteksPengunjung {
	ua := r.UserAgent()
	referer := r.Referer()

	// Ekstrak alamat IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// Ambil IP pertama jika X-Forwarded-For berisi rantai proksi (misal "client, proxy1, proxy2")
		bagian := strings.Split(ip, ",")
		ip = strings.TrimSpace(bagian[0])
	}
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	return AnalisisPermintaan(ua, referer, ip, negaraInput)
}
