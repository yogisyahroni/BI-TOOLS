# Panduan Fitur Data Pipeline (InsightEngine)

Fitur **Data Pipeline** pada InsightEngine memungkinkan Anda untuk mengekstrak, mentransformasi, dan memuat data (ETL/ELT) dari berbagai sumber ke dalam sistem atau tujuan lain secara otomatis, terjadwal, dan terpantau.

---

## 🏗️ Arsitektur & Konsep

Setiap Pipeline terdiri dari 5 komponen utama:

1. **Source (Sumber Data)**: Tempat data berasal.
    * Mendukung: PostgreSQL, MySQL, REST API, dan CSV.
2. **Extraction (Ekstraksi)**: Proses mengambil data dari sumber.
    * Bisa berupa query SQL (`SELECT * FROM ...`) atau request API.
3. **Transformation (Transformasi)**: Memodifikasi data sebelum dimuat.
    * Operasi: Filter, Rename Column, Cast Type, Masking, dll.
4. **Quality Rules (Aturan Kualitas)**: Memastikan data valid sebelum diproses lebih lanjut.
    * Contoh: Pastikan kolom `email` tidak kosong, atau `age` > 0.
5. **Destination (Tujuan)**: Tempat data disimpan.
    * Default: Internal Raw Storage (InsightEngine).
    * Support eksternal DB (future).

---

## 🚀 Tutorial: Membuat Pipeline Baru

Ikuti langkah-langkah berikut untuk membuat pipeline data pertama Anda:

### Langkah 1: Buka Menu Pipeline

1. Klik menu **"Pipelines"** di sidebar kiri.
2. Klik tombol **"+ New Pipeline"** di pojok kanan atas.

### Langkah 2: Konfigurasi Dasar

Isi informasi dasar pipeline:

* **Name**: Beri nama yang jelas (contoh: `Sales Data Daily Sync`).
* **Description**: Opsional, tapi disarankan (contoh: `Sinkronisasi data penjualan harian dari DB Toko`).
* **Schedule (Cron)**: Tentukan jadwal eksekusi otomatis.
  * Contoh: `0 0 * * *` (Setiap hari jam 00:00).
  * Biarkan kosong jika ingin dijalankan manual saja.

### Langkah 3: Konfigurasi Source (Sumber)

Pilih tipe sumber data Anda:

#### Opsi A: Database (Postgres/MySQL)

1. Pilih **Source Type**: `POSTGRES` atau `MYSQL`.
2. **Connection**: Pilih koneksi database yang sudah tersimpan, atau isi detail koneksi (Host, Port, User, Pass) secara manual.
3. **Query**: Masukkan query SQL untuk mengambil data.
    * Contoh: `SELECT id, transaction_date, amount, customer_id FROM sales WHERE transaction_date >= NOW() - INTERVAL '1 day'`

#### Opsi B: REST API

1. Pilih **Source Type**: `REST_API`.
2. **URL**: Masukkan endpoint API (contoh: `https://api.stripe.com/v1/charges`).
3. **Method**: `GET` atau `POST`.
4. **Headers**: Tambahkan header autentikasi jika perlu (contoh: `Authorization: Bearer sk_test_...`).
5. **Body**: Payload JSON jika method POST.

#### Opsi C: CSV

1. Pilih **Source Type**: `CSV`.
2. Upload file CSV atau masukkan path file jika di server.

### Langkah 4: Transformasi (Opsional)

Anda bisa menambahkan langkah transformasi data:

1. Klik **"+ Add Step"** di bagian Transformation.
2. Pilih tipe transformasi:
    * **FILTER**: Saring baris data (contoh: `amount > 100`).
    * **RENAME**: Ubah nama kolom (contoh: `trx_date` -> `transaction_date`).
    * **CAST**: Ubah tipe data (contoh: `amount` (string) -> `amount` (float)).
    * **MASK**: Sensor data sensitif (contoh: mask email/ktp).

### Langkah 5: Quality Rules (Validasi Data)

Tambahkan aturan untuk menjaga kualitas data. Jika aturan dilanggar, sistem akan mencatatnya sebagai "Quality Violation".

1. Klik **"+ Add Rule"**.
2. Pilih kolom yang akan divalidasi.
3. Pilih tipe rule:
    * **NOT_NULL**: Kolom tidak boleh kosong.
    * **UNIQUE**: Nilai harus unik.
    * **RANGE**: Nilai harus dalam rentang tertentu (min/max).
    * **REGEX**: Nilai harus cocok dengan pola regex (misal format email).
4. **Severity**:
    * `WARN`: Pipeline tetap lanjut, cuma dicatat log warning.
    * `FAIL`: Eksekusi pipeline akan **dibatalkan** jika aturan ini dilanggar.

### Langkah 6: Simpan

Klik **"Create Pipeline"** untuk menyimpan konfigurasi.

---

## ▶️ Menjalankan & Monitoring Pipeline

### Cara Menjalankan

* **Manual**: Di halaman detail pipeline, klik tombol **"Run Now"**.
* **Otomatis**: Pipeline akan jalan sesuai jadwal Cron yang Anda set.

### Monitoring Eksekusi

Di halaman detail pipeline, lihat tab **"History"** atau **"Executions"**:

* **Status**: `PENDING` -> `PROCESSING` -> `COMPLETED` / `FAILED`.
* **Metrics**: Lihat jumlah baris (`Rows Processed`) dan ukuran data (`Bytes Processed`).
* **Logs**: Klik salah satu baris history untuk melihat log detail per step (Extract, Transform, Load).

### Troubleshooting

Jika status **FAILED**:

1. Buka detail eksekusi tersebut.
2. Lihat bagian **Logs**. Cari baris berwarna merah (ERROR).
3. Contoh error umum:
    * *Connection Refused*: Cek kredensial database source.
    * *SQL Syntax Error*: Cek query SQL di konfigurasi source.
    * *Quality Rule Failure*: Data kotor melanggar rule dengan severity `FAIL`.

---

## 🛠️ Fitur Lanjutan

### Data Lineage (Asal Usul Data)

Setiap eksekusi mencatat dari mana data berasal dan transformasi apa yang terjadi. Ini berguna untuk audit trail.

### Notifikasi (Coming Soon)

Integrasi notifikasi ke Email/Slack jika pipeline gagal atau ditemukan anomali data.
