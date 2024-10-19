# Gunakan image Go resmi sebagai base image
FROM golang:1.23.1 AS builder

# Set working directory
WORKDIR /app

# Copy go mod dan sum
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh kode sumber
COPY . .

# Build aplikasi
RUN go build -o main .

# Gunakan image minimal untuk menjalankan aplikasi
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary dari stage builder
COPY --from=builder /app/main .

# Expose port (ganti dengan port yang digunakan aplikasi Anda)
EXPOSE 8080

# Jalankan aplikasi
CMD ["./main"]
