# Bookshelf App

REST API untuk mengelola data **buku** dan **kategori**-nya. Dibuat dengan Go (Gin), PostgreSQL, migrasi menggunakan `sql-migrate`, dan autentikasi **JSON Web Token (JWT)**.

## Tech Stack

- [Gin](https://github.com/gin-gonic/gin): HTTP framework
- [lib/pq](https://github.com/lib/pq): driver PostgreSQL
- [sql-migrate](https://github.com/rubenv/sql-migrate): migrasi database
- [golang-jwt](https://github.com/golang-jwt/jwt): token JWT
- bcrypt: hashing password user

## Struktur Folder

```
bookshelf-app/
├── config/          # koneksi database & helper env
├── controllers/     # handler untuk auth, kategori, buku
├── database/
│   ├── migrate.go   # menjalankan migrasi (file SQL di-embed ke binary)
│   └── migrations/  # file migrasi SQL
├── middlewares/     # middleware JWT & pembuatan token
├── models/          # struct data & logic konversi thickness
├── routes/          # definisi path API
├── main.go
├── Dockerfile
└── .env.example
```

## Cara Menjalankan (Lokal)

1. Buat database PostgreSQL:
   ```bash
   createdb bookshelf_db
   ```
2. Salin file env lalu sesuaikan isinya:
   ```bash
   cp .env.example .env
   ```
3. Jalankan aplikasi:
   ```bash
   go run .
   ```
   Migrasi berjalan otomatis setiap kali aplikasi start, jadi tabel `users`, `categories`, `books` beserta user default langsung dibuat.

Server berjalan di `http://localhost:8080`.

### Environment Variable

| Variable       | Keterangan                                                 | Default             |
|----------------|------------------------------------------------------------|---------------------|
| `DATABASE_URL` | Connection string PostgreSQL (Neon / Railway). Jika diisi, `DB_*` diabaikan | -                   |
| `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE` | Dipakai jika `DATABASE_URL` kosong | `localhost`, `5432`, `postgres`, kosong, `bookshelf_db`, `disable` |
| `JWT_SECRET`   | Secret untuk menandatangani token                          | `rahasia-bookshelf` |
| `PORT`         | Port server                                                | `8080`              |

## Koneksi ke Neon (PostgreSQL Cloud)

1. Buat project di [neon.tech](https://neon.tech), lalu buka **Connect** dan copy connection string-nya.
2. Isi ke `.env`:
   ```env
   DATABASE_URL=postgresql://user:password@ep-xxxx.region.aws.neon.tech/neondb?sslmode=require
   ```
3. Jalankan `go run .`. Tabel dan user default dibuat otomatis oleh migrasi.

> Connection string dari Neon biasanya berakhiran `&channel_binding=require`. Parameter ini tidak didukung driver `lib/pq`, jadi aplikasi otomatis membuangnya. Koneksi tetap terenkripsi lewat `sslmode=require`.

## Autentikasi

Semua path `/api/categories` dan `/api/books` wajib memakai token JWT.

User default (dibuat lewat migrasi):

| username | password   |
|----------|------------|
| `admin`  | `admin123` |

**1. Login untuk mendapatkan token**

```bash
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

Response:
```json
{ "message": "login berhasil", "token": "eyJhbGciOi..." }
```

**2. Kirim token di header setiap request**

```
Authorization: Bearer <token>
```

Token berlaku 24 jam. Username dari token otomatis disimpan ke kolom `created_by` / `modified_by`.

## Daftar Path API

| Method | Path                         | Keterangan                              | Auth |
|--------|------------------------------|-----------------------------------------|------|
| POST   | `/api/users/login`           | Login, mendapatkan token JWT            | -    |
| GET    | `/api/categories`            | Menampilkan seluruh kategori            | JWT  |
| POST   | `/api/categories`            | Menambahkan kategori                    | JWT  |
| GET    | `/api/categories/:id`        | Menampilkan detail kategori             | JWT  |
| PUT    | `/api/categories/:id`        | Mengubah kategori                       | JWT  |
| DELETE | `/api/categories/:id`        | Menghapus kategori                      | JWT  |
| GET    | `/api/categories/:id/books`  | Menampilkan buku berdasarkan kategori   | JWT  |
| GET    | `/api/books`                 | Menampilkan seluruh buku                | JWT  |
| POST   | `/api/books`                 | Menambahkan buku                        | JWT  |
| GET    | `/api/books/:id`             | Menampilkan detail buku                 | JWT  |
| PUT    | `/api/books/:id`             | Mengubah buku                           | JWT  |
| DELETE | `/api/books/:id`             | Menghapus buku                          | JWT  |

### Body Kategori (POST / PUT)

```json
{ "name": "Novel" }
```

### Body Buku (POST / PUT)

```json
{
  "title": "Laskar Pelangi",
  "description": "Novel karya Andrea Hirata",
  "image_url": "https://example.com/laskar-pelangi.jpg",
  "release_year": 2005,
  "price": 85000,
  "total_page": 529,
  "category_id": 1
}
```

Kolom `thickness` **tidak** dikirim dari JSON, tetapi dihitung otomatis dari `total_page`:

| total_page | thickness |
|------------|-----------|
| > 100      | `tebal`   |
| ≤ 100      | `tipis`   |

## Validasi

| Kasus                                                  | Status | Pesan error                                                    |
|--------------------------------------------------------|--------|----------------------------------------------------------------|
| Tanpa token / token tidak valid                        | 401    | `token tidak ditemukan ...` / `token tidak valid atau sudah kedaluwarsa` |
| Hapus / update / detail kategori yang tidak ada        | 404    | `kategori dengan ID <id> tidak ditemukan`                      |
| Hapus / update / detail buku yang tidak ada            | 404    | `buku dengan ID <id> tidak ditemukan`                          |
| `release_year` di luar 1980–2024                       | 400    | `release_year harus di antara 1980 sampai 2024`                |
| `category_id` buku tidak ada                           | 404    | `kategori dengan ID <id> tidak ditemukan`                      |
| Hapus kategori yang masih punya buku                   | 409    | `kategori masih digunakan oleh buku, hapus bukunya terlebih dahulu` |
| `title` / `name` kosong, `total_page` ≤ 0, `price` < 0 | 400    | pesan sesuai field                                             |

## Database Design

**books**: `id`, `title`, `description`, `image_url`, `release_year`, `price`, `total_page`, `thickness`, `category_id` (FK → categories.id), `created_at`, `created_by`, `modified_at`, `modified_by`

**categories**: `id`, `name`, `created_at`, `created_by`, `modified_at`, `modified_by`

**users**: `id`, `username`, `password`, `created_at`, `created_by`, `modified_at`, `modified_by`

> Catatan: di dokumen desain, `modified_by` tertulis `timestamp`. Di project ini kolom tersebut memakai `varchar` karena isinya username yang mengubah data, sama seperti `created_by`.

## Deploy ke Railway

1. Push repository ini ke GitHub.
2. Di Railway: **New Project → Deploy from GitHub repo**, pilih repo ini (Railway akan build memakai `Dockerfile`).
3. Siapkan database, pilih salah satu:
   - **Neon**: pakai connection string dari Neon.
   - **PostgreSQL Railway**: **New → Database → PostgreSQL**.
4. Di service aplikasi, buka **Variables** lalu tambahkan:
   - `DATABASE_URL` = connection string Neon, atau `${{Postgres.DATABASE_URL}}` jika memakai PostgreSQL Railway
   - `JWT_SECRET` = string acak yang panjang
5. Buka **Settings → Networking → Generate Domain** untuk mendapatkan URL publik.

Migrasi berjalan otomatis saat aplikasi start, jadi tidak perlu menjalankan SQL manual.
