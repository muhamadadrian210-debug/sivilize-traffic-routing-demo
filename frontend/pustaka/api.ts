// Mendapatkan URL dasar backend dari konfigurasi lingkungan
export function dapatkanUrlApi(): string {
  return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
}
