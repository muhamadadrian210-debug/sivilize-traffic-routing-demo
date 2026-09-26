// Mendapatkan URL dasar backend dari konfigurasi lingkungan
export function dapatkanUrlApi(): string {
  return process.env.NEXT_PUBLIC_API_URL || "https://democrats-parish-press-unlock.trycloudflare.com";
}
