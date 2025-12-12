# Setup Server Baru untuk Blue-Green Deployment Go App

Panduan ini akan membantu Anda menyiapkan server baru (misalnya, VPS atau Bare Metal dengan Ubuntu 22.04/24.04 LTS) agar siap menerima deployment otomatis dari GitHub Actions CI/CD yang telah kita buat, termasuk Blue-Green Deployment.

## Prasyarat

*   Akses SSH ke server sebagai user `root`.
*   (Opsional, tapi sangat disarankan) Domain Anda sudah diarahkan ke alamat IP public server.

---

## Langkah 1: Update OS & Security Dasar

Login ke server Anda via SSH sebagai `root` dan jalankan perintah berikut:

```bash
# Update repository dan paket sistem
apt update && apt upgrade -y

# Install tools dasar yang diperlukan
apt install -y git curl wget ufw fail2ban

# Setup Firewall (UFW)
# Izinkan koneksi SSH agar Anda tidak terputus
ufw allow OpenSSH

# Izinkan HTTP dan HTTPS untuk akses web
ufw allow 80/tcp  # HTTP
ufw allow 443/tcp # HTTPS

# Aktifkan firewall
ufw enable
# Jika diminta konfirmasi, ketik 'y' dan tekan Enter
```

---

## Langkah 2: Buat User Khusus Deployment

Sangat tidak disarankan menggunakan user `root` untuk deployment otomatis. Kita akan membuat user baru bernama `deploy` dengan hak akses yang sesuai.

```bash
# Buat user baru bernama 'deploy'
adduser deploy

# (Opsional) Beri user 'deploy' akses sudo
# Ini berguna untuk debugging manual atau jika script deployment membutuhkan hak sudo
usermod -aG sudo deploy

# Beralih ke user 'deploy'
su - deploy
```

---

## Langkah 3: Install Docker & Docker Compose

Instalasi Docker Engine dan Docker Compose CLI diperlukan untuk menjalankan container aplikasi. Lakukan ini sebagai user `deploy` (Anda akan diminta password sudo jika `deploy` adalah user sudo).

```bash
# 1. Setup repository Docker resmi
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch=\"$(dpkg --print-architecture)\" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  \"$(. /etc/os-release && echo \"$VERSION_CODENAME\")\" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 2. Install Docker Engine, CLI, Containerd, dan Docker Compose Plugin
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 3. Masukkan user 'deploy' ke group 'docker'
# Ini PENTING agar script CI/CD dapat menjalankan perintah Docker tanpa perlu sudo
sudo usermod -aG docker "$USER"

# 4. Aktifkan perubahan group tanpa logout/login (opsional)
# Perintah ini mengaktifkan group baru untuk sesi shell saat ini
newgrp docker

# 5. Verifikasi instalasi Docker
docker ps
# Anda seharusnya tidak melihat error 'permission denied'. Jika ada, pastikan langkah 3 dan 4 sudah benar.
```

---

## Langkah 4: Setup SSH Key untuk GitHub Actions

Kita akan membuat pasangan SSH key di server ini, lalu public key-nya akan ditambahkan ke GitHub sebagai Deploy Key, dan private key-nya akan disimpan di GitHub Actions Secrets.

**Di Server (masih sebagai user `deploy`):**

```bash
# Generate pasangan SSH Key baru. Tekan Enter untuk default path dan kosongkan passphrase.
ssh-keygen -t ed25519 -C "github-actions-for-server-deployment"

# Tambahkan public key ke authorized_keys agar user ini bisa login via SSH Key (opsional, tapi baik untuk consistency)
cat ~/.ssh/id_ed25519.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# Tampilkan PRIVATE KEY Anda. COPY SELURUH OUTPUT INI! Ini akan digunakan di GitHub Secrets.
cat ~/.ssh/id_ed25519
```

**Di GitHub Repository Anda (Buka: `Settings` -> `Secrets and variables` -> `Actions`):**

Buat Repository Secrets baru:

1.  **`PROD_HOST`**: Alamat IP public server Production Anda (misal: `192.0.2.1`).
2.  **`PROD_USER`**: `deploy`.
3.  **`PROD_KEY`**: Paste **seluruh isi Private Key** dari output `cat ~/.ssh/id_ed25519` di atas.
4.  **`STAGING_HOST`**: Alamat IP public server Staging Anda (jika terpisah, jika sama dengan Production, gunakan IP yang sama).
5.  **`STAGING_USER`**: `deploy`.
6.  **`STAGING_KEY`**: Paste **seluruh isi Private Key** dari output `cat ~/.ssh/id_ed25519` di atas (atau private key dari server staging jika terpisah).

   *Catatan*: Jika Anda menggunakan server Production dan Staging yang sama, Anda bisa menggunakan `PROD_HOST` = `STAGING_HOST`, `PROD_USER` = `STAGING_USER`, dan `PROD_KEY` = `STAGING_KEY`. Namun, untuk lingkungan nyata, disarankan server yang terpisah.

---

## Langkah 5: Setup Deploy Key untuk Pull Repository

Agar server dapat melakukan `git pull` dari repository GitHub Anda (terutama jika private repository).

**Di Server (masih sebagai user `deploy`):**

```bash
# Tampilkan PUBLIC KEY Anda. COPY SELURUH OUTPUT INI! Ini akan digunakan sebagai Deploy Key di GitHub.
cat ~/.ssh/id_ed25519.pub
```

**Di GitHub Repository Anda (Buka: `Settings` -> `Deploy keys`):**

1.  Klik tombol **Add deploy key**.
2.  **Title**: Masukkan nama deskriptif (misal: `Production/Staging Server`).
3.  **Key**: Paste **seluruh isi Public Key** dari output `cat ~/.ssh/id_ed25519.pub` di atas.
4.  **Allow write access**: Biarkan kotak ini **TIDAK DICENTANG**. Server hanya perlu membaca kode (read-only), tidak menulis ke repository.

**Di Server (Test koneksi SSH ke GitHub):**

```bash
ssh -T git@github.com
# Jika diminta konfirmasi fingerprint, ketik 'yes' dan tekan Enter.
# Anda harus melihat pesan yang mirip dengan:
# "Hi <username>! You've successfully authenticated, but GitHub does not provide shell access."
```

---

## Langkah 6: Siapkan Direktori Aplikasi

Sesuai dengan konfigurasi `cd.yml` kita, aplikasi akan di-deploy ke `/app/production` dan `/app/staging`. Kita perlu menyiapkan direktori ini dan mengkloning repository Anda di sana.

**Di Server (sebagai user `deploy`):**

```bash
# Kembali ke direktori home user deploy
cd ~

# Buat direktori /app (perlu sudo karena di root)
sudo mkdir -p /app/production
sudo mkdir -p /app/staging

# Ubah kepemilikan direktori /app ke user 'deploy'
sudo chown -R deploy:deploy /app

# Pindah ke direktori produksi
cd /app/production

# Kloning repository aplikasi Anda
# Pastikan URL sesuai dengan repository Anda (misal: Roisfaozi/go-clean-boilerplate)
git clone git@github.com:Roisfaozi/go-clean-boilerplate.git .

# Jika staging di server yang sama, pindah ke direktori staging
cd /app/staging

# Kloning repository aplikasi Anda
git clone git@github.com:Roisfaozi/go-clean-boilerplate.git .
```

---

## Langkah 7: Konfigurasi Environment Variable Aplikasi (`.env`)

Aplikasi Anda membutuhkan environment variable (misalnya, kredensial database). Meskipun kita juga mengonfigurasi di `docker-compose.prod.yml`, membuat file `.env` di server adalah praktik baik untuk manajemen rahasia dan konfigurasi spesifik lingkungan. Docker Compose akan otomatis membaca file `.env` jika ada di direktori yang sama.

**Untuk Lingkungan Produksi (`/app/production/.env`):**

```bash
cd /app/production
nano .env
```

**Isi file `.env` (Sesuaikan dengan kredensial Production Anda):**

```ini
# Database Credentials
MYSQL_HOST=mysql_prod
MYSQL_PORT=3306
MYSQL_DATABASE=gin_starter_prod
MYSQL_USER=prod_user
MYSQL_PASSWORD=ProdStrongPasswordPleaseChangeMe!

# Redis Credentials
REDIS_HOST=redis_prod:6379
REDIS_PASSWORD=YourRedisProdPassword

# Application Config
APP_ENV=production
GIN_MODE=release
PORT=8080

# JWT Secrets (Generate these securely)
JWT_ACCESS_SECRET=your_super_secret_access_key_for_production
JWT_REFRESH_SECRET=your_super_secret_refresh_key_for_production
```

**Untuk Lingkungan Staging (`/app/staging/.env`):**

```bash
cd /app/staging
nano .env
```

**Isi file `.env` (Sesuaikan dengan kredensial Staging Anda):**

```ini
# Database Credentials
MYSQL_HOST=mysql_prod # Atau mysql_staging jika terpisah
MYSQL_PORT=3306
MYSQL_DATABASE=gin_starter_staging
MYSQL_USER=staging_user
MYSQL_PASSWORD=StagingStrongPassword!

# Redis Credentials
REDIS_HOST=redis_prod # Atau redis_staging jika terpisah
REDIS_PASSWORD=YourRedisStagingPassword

# Application Config
APP_ENV=staging
GIN_MODE=debug # Atau release
PORT=8080

# JWT Secrets (Generate these securely)
JWT_ACCESS_SECRET=your_super_secret_access_key_for_staging
JWT_REFRESH_SECRET=your_super_secret_refresh_key_for_staging
```
*Catatan:* Pastikan `MYSQL_HOST` dan `REDIS_HOST` sesuai dengan nama service di `docker-compose.prod.yml` jika Anda menggunakan Docker Compose untuk database dan Redis.

---

## Langkah 8: Bootstrap Awal (First Run Deployment)

Sebelum CI/CD dapat berjalan lancar, kita perlu menyalakan infrastruktur dasar di server (database, redis, Nginx) dan melakukan deployment awal secara manual. Ini memastikan semua volume dan konfigurasi awal terbentuk.

**Di Direktori Proyek Production (`/app/production`):**

1.  **Pindah ke direktori produksi:**
    ```bash
    cd /app/production
    ```
2.  **Bangun dan jalankan service database, redis, dan Nginx:**
    ```bash
    docker compose -f docker-compose.prod.yml up -d mysql_prod redis_prod nginx
    ```
    *   Perintah ini akan membuat container database, Redis, dan Nginx berjalan di background.
    *   Volume Docker juga akan terinisialisasi.
3.  **Inisialisasi `upstream.conf`:**
    Pastikan file `deploy/nginx/upstream.conf` menunjuk ke `app-blue` sebagai default awal.
    ```bash
    echo "upstream backend { server app-blue:8080; }" > deploy/nginx/upstream.conf
    ```
4.  **Lakukan deployment awal untuk `app-blue`:**
    ```bash
    docker compose -f docker-compose.prod.yml up -d --build app-blue
    ```
    Ini akan membangun image aplikasi dari `deploy/Dockerfile` dan menjalankan container `app-blue`.

Verifikasi:
Cek apakah semua container berjalan:
```bash
docker ps
```
Anda seharusnya melihat `nginx_gateway`, `app-blue`, `mysql_prod`, dan `redis_prod` dalam status `Up` dan `healthy`.

---

## Selesai!

Server Anda sekarang sudah siap menerima deployment otomatis.

```