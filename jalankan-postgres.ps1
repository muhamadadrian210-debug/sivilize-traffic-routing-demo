# Skrip untuk menjalankan PostgreSQL portable di lingkungan lokal
# Jalankan sekali sebelum menghidupkan server backend
$pgBin = "$env:USERPROFILE\postgresql17\pgsql\bin"
$pgData = "$env:USERPROFILE\postgresql17\data"
$pgLog = "$env:USERPROFILE\postgresql17\postgres.log"

$status = & "$pgBin\pg_ctl.exe" -D $pgData status 2>&1
if ($status -match "server is running") {
    Write-Host "PostgreSQL sudah berjalan."
} else {
    Write-Host "Menghidupkan PostgreSQL portable..."
    & "$pgBin\pg_ctl.exe" -D $pgData -l $pgLog start
    Start-Sleep -Seconds 2
    Write-Host "PostgreSQL berjalan. Log: $pgLog"
}
