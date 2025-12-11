# Stage 1: Builder
# Menggunakan image Go resmi sebagai base untuk tahap build.
# Kita pilih alpine untuk ukuran image yang lebih kecil.
FROM golang:1.21-alpine AS builder

# Menentukan working directory di dalam container
WORKDIR /app

# Menginstal dependensi build yang diperlukan, seperti git untuk 'go get' dan cgo.
# Walaupun kita akan membangun dengan CGO_ENABLED=0, ada baiknya memiliki dependensi dasar
# jika ada kebutuhan lain di masa depan.
# Namun, untuk CGO_ENABLED=0, biasanya tidak perlu gcc atau libc-dev di tahap builder,
# tetapi go mod download mungkin masih memerlukan git.
RUN apk add --no-cache git

# Copy go.mod dan go.sum terlebih dahulu untuk memanfaatkan Docker layer caching.
# Jika dependensi tidak berubah, Docker tidak perlu mendownload ulang.
COPY go.mod .
COPY go.sum .

# Download semua dependensi.
# Menggunakan GO111MODULE=on memastikan Go Modules digunakan.
RUN GO111MODULE=on go mod download

# Copy sisa kode sumber aplikasi
COPY . .

# Build aplikasi
# CGO_ENABLED=0: Membangun binary yang statis dan tidak bergantung pada pustaka C,
#                menghasilkan image runner yang sangat kecil.
# GOOS=linux: Menentukan OS target.
# GOARCH=amd64: Menentukan arsitektur target.
# -ldflags="-s -w": Mengurangi ukuran binary dengan menghilangkan tabel simbol (-s)
#                   dan informasi debugging DWARF (-w).
# -o /app/bin/api: Menentukan nama dan lokasi binary yang akan dihasilkan.
# ./cmd/api: Path ke paket main aplikasi Anda.
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -ldflags="-s -w" -o /app/bin/api ./cmd/api

# ---

# Stage 2: Runner
# Menggunakan image yang sangat minimal (alpine) untuk ukuran final yang sekecil mungkin.
FROM alpine:latest

# Menentukan working directory di dalam container
WORKDIR /app

# Instal sertifikat CA jika aplikasi Anda melakukan panggilan HTTPS eksternal
# atau berinteraksi dengan layanan lain yang menggunakan sertifikat SSL/TLS.
# Ini sangat umum dan direkomendasikan.
RUN apk add --no-cache ca-certificates

# Copy binary yang sudah terkompilasi dari tahap builder.
COPY --from=builder /app/bin/api /usr/local/bin/api

# Copy file konfigurasi Casbin
COPY internal/config/casbin_model.conf /app/internal/config/casbin_model.conf

# Copy folder migrasi database.
# Ini penting jika aplikasi Anda menjalankan migrasi saat startup (walaupun tidak direkomendasikan di prod).
# Atau jika Anda memiliki mekanisme external migration runner.
COPY db/migrations /app/db/migrations

# Copy folder dokumentasi Swagger.
# Ini diperlukan jika aplikasi Go Anda yang menyajikan Swagger UI secara langsung.
# Jika Swagger UI disajikan oleh web server terpisah (misalnya Nginx), folder ini tidak perlu dicopy.
COPY docs /app/docs

# Expose port yang digunakan aplikasi Anda (misalnya 8080).
# Ini hanya mendokumentasikan port yang digunakan, tidak benar-benar mempublikasikannya.
EXPOSE 8080

# Set environment variables for the application if needed,
# though ideally these are passed via docker run or docker-compose.
# ENV APP_PORT=8080
# ENV DB_HOST=...
# ENV REDIS_HOST=...

# Mendefinisikan ENTRYPOINT untuk menjalankan aplikasi.
# Menggunakan array JSON adalah format yang direkomendasikan karena tidak menggunakan shell
# dan menghindari masalah sinyal proses.
ENTRYPOINT ["/usr/local/bin/api"]

# CMD adalah argumen default untuk ENTRYPOINT.
# Dalam kasus ini, binary Go akan langsung dieksekusi tanpa argumen tambahan.
# Anda bisa menambahkan argumen di sini jika binary Anda menerimanya.
# CMD ["serve"]
