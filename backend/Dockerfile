# Tahap 1: Kompilasi biner Go
FROM golang:1.23-alpine AS pembangun

WORKDIR /app

# Salin modul dan unduh dependensi
COPY go.mod go.sum ./
RUN go mod download

# Salin seluruh kode backend
COPY . .

# Kompilasi aplikasi server statis
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Tahap 2: Runtime image minimal
FROM alpine:3.20

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

COPY --from=pembangun /app/server .

# Port default
EXPOSE 8080

CMD ["./server"]
