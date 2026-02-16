# Panduan Data Lineage & Governance

Fitur **Data Lineage & Governance** di InsightEngine dirancang untuk memberikan transparansi penuh atas aliran data Anda, sekaligus memastikan keamanan dan kepatuhan terhadap regulasi privasi (seperti UU PDP atau GDPR).

---

## 🔍 Data Lineage (Pelacakan Alur Data)

Data Lineage bukan sekadar gambar, tapi peta otomatis yang menjelaskan perjalanan data Anda.

### Apa Fungsinya?

1. **Analisis Dampak (Impact Analysis)**
    * *Pertanyaan*: "Jika saya mengubah kolom `customer_id` di tabel database, dashboard apa saja yang akan error?"
    * *Jawaban*: Lineage akan menyoroti semua Query dan Dashboard yang bergantung pada tabel tersebut.

2. **Analisis Akar Masalah (Root Cause Analysis)**
    * *Pertanyaan*: "Kenapa angka 'Total Revenue' di dashboard CEO turun drastis?"
    * *Jawaban*: Anda bisa menelusuri mundur (trace back) dari Dashboard -> Query -> Tabel -> DataSource. Mungkin ada masalah di data mentah.

3. **Visualisasi Alur Otomatis**
    * Sistem secara otomatis membaca SQL Query Anda dan membangun grafik:
    * `DataSource` ➡ `Tabel` ➡ `Saved Query` ➡ `Dashboard`

---

## 🛡️ Data Governance (Tata Kelola & Keamanan)

Fitur ini memastikan data sensitif tetap aman dan organisasi Anda patuh hukum.

### 1. Klasifikasi Data & PII (Personally Identifiable Information)

Anda bisa menandai kolom tertentu sebagai data sensitif.

* **Tipe**: Email, NIK, No. Telepon, Credit Card, dll.
* **Manfaat**: Memudahkan audit data sensitif.

### 2. Data Masking (Penyensoran Otomatis)

Melindungi data dari mata yang tidak berhak tanpa mengubah data aslinya di database.

* **Dynamic Masking**: Data di-masking *on-the-fly* saat di-query.
* **Metode Masking**:
  * `Full`: `*****` (Disembunyikan total)
  * `Partial`: `123-***-789` (Tampilkan sebagian)
  * `Email`: `j***@example.com`
  * `None`: Tampilkan apa adanya (untuk Admin/User berizin).

### 3. Column-Level Security

Mengatur siapa yang boleh melihat kolom tertentu.

* Contoh: Tim Marketing boleh lihat `Email` customer, tapi tidak boleh lihat `Gaji`. Tim HR boleh lihat `Gaji`.

### 4. Compliance (Kepatuhan Regulasi)

Mendukung kepatuhan terhadap UU Pelindungan Data Pribadi (PDP) / GDPR.

* **Right to Erasure**: Mekanisme untuk menghapus total data user tertentu jika diminta (hak untuk dilupakan).
* **Audit Logs**: Mencatat - siapa mengakses data apa, dan kapan.

---

## 💡 Kesimpulan: Kenapa Ini Penting?

| Fitur | Manfaat Bisnis |
| :--- | :--- |
| **Lineage** | Menghemat waktu tim data saat debugging error (dari berjam-jam jadi hitungan menit). |
| **Governance** | Mencegah kebocoran data sensitif (customer trust) dan denda regulasi (compliance). |
