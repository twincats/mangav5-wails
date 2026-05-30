# Pengaturan Scraping di Halaman Settings

Halaman **Settings** digunakan untuk:

- Menambahkan aturan scraping (site rule) untuk berbagai website manga
- Menguji apakah aturan sudah bekerja dengan benar
- Menyimpan dan memuat kembali aturan yang sudah dibuat
- Mengatur folder utama untuk penyimpanan file manga

Dokumen ini menjelaskan cara penggunaan halaman **Settings** untuk pengguna akhir, tanpa perlu memahami detail teknis di belakangnya.

---

## 1. Tampilan Utama Halaman Settings

Halaman Settings terbagi menjadi beberapa bagian:

- Baris atas (toolbar cepat)
- Baris kedua (informasi site rule)
- Area editor (aturan Manga dan Chapter serta hasil uji)
- Modal **Load Rules** (memilih rule yang sudah tersimpan)
- Dialog pengaturan folder download manga

Setiap baris dijelaskan di bawah.

---

## 2. Toolbar Atas – Uji Rule dan Simpan

Di bagian paling atas terdapat:

1. **Manga Rule URL + tombol GO**
   - Kolom ini untuk memasukkan URL halaman manga (bukan URL chapter).
   - Contoh: halaman detail manga yang berisi daftar chapter.
   - Setelah mengisi, klik **GO** di sampingnya.
   - Aplikasi akan mencoba mengambil informasi manga sesuai rule yang ada di tab **Manga Rule**.
   - Hasil uji akan ditampilkan di panel **Scrape Result**.

2. **Chapter Rule URL + tombol GO**
   - Kolom ini untuk memasukkan URL halaman chapter.
   - Contoh: halaman baca satu chapter yang berisi daftar gambar halaman.
   - Setelah mengisi, klik **GO** di sampingnya.
   - Aplikasi akan mencoba mengambil daftar halaman gambar sesuai rule yang ada di tab **Chapter Rule**.
   - Hasil uji juga akan ditampilkan di panel **Scrape Result**.

3. **Tombol Download (test)**
   - Tombol **Download** di toolbar Settings digunakan sebagai **uji coba cepat**, bukan untuk download manga utama.
   - Tombol ini hanya aktif jika hasil di **Scrape Result** berisi daftar halaman yang valid.
   - Saat ditekan, aplikasi mencoba mengunduh gambar berdasarkan hasil tersebut ke lokasi uji (untuk kebutuhan testing).
   - Lokasi uji mengikuti konfigurasi **Manga Directory**:
     - Output: `<Manga Directory>/__settings_test/<site_key>/...`
     - Jika Manga Directory belum di-set, tombol ini akan meminta Anda mengatur folder terlebih dulu.

4. **Tombol Save Rules**
   - Digunakan untuk menyimpan rule yang sedang Anda kerjakan ke database.
   - Tombol hanya aktif jika:
     - Informasi situs (site_key, name, domains) sudah terisi,
     - Aturan Manga dan Chapter sudah diisi,
     - Dan keduanya dinyatakan valid oleh aplikasi.
   - Setelah disimpan, rule bisa digunakan di halaman **Download**.

5. **Tombol Load Rules**
   - Membuka modal berisi daftar rule yang sudah pernah disimpan.
   - Saat Anda memilih salah satu rule, semua kolom dan editor akan terisi otomatis sesuai rule tersebut.

---

## 3. Baris Kedua – Informasi Site Rule

Di baris kedua terdapat beberapa kolom:

1. **site_key**
   - Kode unik untuk membedakan satu situs dengan situs lain.
   - Biasanya diisi otomatis berdasarkan nama situs pada rule.
   - Contoh: `mangadomain`, `asmotoon`, dll.

2. **name**
   - Nama situs yang lebih mudah dibaca.
   - Biasanya juga diisi otomatis dari rule.
   - Contoh: `Asmotoon`, `MangaDex`, dll.

3. **domains**
   - Daftar domain yang terkait dengan situs tersebut.
   - Digunakan aplikasi untuk menebak rule mana yang harus dipakai saat Anda memasukkan URL di halaman Download.
   - Format umumnya berupa array teks, misalnya: `["asmotoon.com"]`.

4. **Switch enabled / disabled**
   - Menyalakan atau mematikan rule.
   - Jika **disabled**, rule ini tidak akan dipakai di halaman Download.

5. **Status Manga Rule dan Chapter Rule**
   - Dua tombol kecil dengan ikon:
     - **Manga Rule**
     - **Chapter Rule**
   - Jika ikon berwarna hijau → rule dianggap valid.
   - Jika merah → masih ada masalah di isi rule (misalnya informasi penting belum diatur).
   - Anda tidak perlu tahu detail error-nya; cukup pastikan ikon sudah hijau sebelum menyimpan.

Di sisi kanan baris kedua terdapat:

1. **Tombol Set Config Download Manga Directory**
   - Membuka dialog untuk mengatur folder utama tempat manga akan disimpan.
   - Masukkan path folder di komputer Anda (misalnya `D:\Manga`).
   - Folder ini akan digunakan oleh fitur download utama di halaman lain.

2. **Tombol Clear (ikon penghapus)**
   - Mengosongkan semua input dan editor.
   - Berguna ketika Anda ingin membuat rule baru dari nol.

---

## 4. Area Editor – Mengelola Aturan dan Melihat Hasil

Area editor terbagi menjadi dua bagian utama:

### 4.1. Tab Manga Rule

- Editor di tab ini berisi aturan bagaimana aplikasi membaca halaman **detail manga**.
- Di sinilah Anda menulis atau menempelkan konfigurasi rule untuk satu situs.
- Anda tidak wajib memahami struktur teknis, cukup mengikuti contoh yang sudah disediakan atau mempercayakan ke pembuat rule.
- Hal yang perlu diperhatikan oleh pengguna:
  - Jika ada kesalahan besar, ikon **Manga Rule** di baris kedua akan merah.
  - Jika semuanya sudah baik, ikon akan hijau dan rule siap dipakai.

### 4.2. Tab Chapter Rule

- Editor ini berisi aturan bagaimana aplikasi membaca halaman **chapter** (daftar gambar per halaman).
- Sama seperti Manga Rule, biasanya diisi oleh pembuat rule.
- Pengguna cukup memastikan:
  - Untuk situs yang akan dipakai download, rule chapter sudah terisi.
  - Ikon **Chapter Rule** di baris kedua sudah hijau.

### 4.3. Tab Scrape Result

- Panel ini menampilkan hasil uji saat Anda klik tombol **GO** di toolbar atas.
- Jika Anda:
  - Mengisi **Manga Rule URL** dan klik GO → di sini akan muncul data detail manga dan daftar chapter.
  - Mengisi **Chapter Rule URL** dan klik GO → di sini akan muncul daftar URL gambar untuk satu chapter.
- Fungsi panel ini untuk:
  - Mengecek apakah rule mengambil data yang masuk akal.
  - Menjadi dasar test untuk tombol **Download** di halaman Settings.

### 4.4. Tab Raw HTML

- Tab **Raw HTML** menampilkan HTML mentah yang didapat saat proses scrape.
- Gunanya terutama untuk pembuat rule:
  - Memastikan selector CSS yang dipakai memang ada di HTML yang diterima aplikasi.
  - Membandingkan HTML aktual vs yang terlihat di browser (beberapa situs merender berbeda jika ada proteksi / lazy-load).
- Catatan:
  - Tidak semua scrape selalu mengembalikan Raw HTML (tergantung strategy dan sumber data).
  - Jika hasil scrape memuat field `__raw_html`, aplikasi akan menyembunyikan field tersebut dari tab **Scrape Result** dan menampilkannya di tab **Raw HTML**.

---

## 5. Modal Load Rules

Saat klik **Load Rules**, akan muncul dialog berisi daftar rule yang sudah tersimpan:

- Setiap baris berisi:
  - Nama rule (name)
  - Satu contoh domain utama
- Klik salah satu rule untuk:
  - Mengisi otomatis:
    - `site_key`, `name`, `domains`
    - Konten **Manga Rule**
    - Konten **Chapter Rule**
  - Menutup modal dan kembali ke halaman utama Settings.

Ini memudahkan Anda:

- Berpindah antar situs tanpa perlu mengetik ulang rule.
- Mengedit rule lama jika situs mengalami perubahan.

---

## 6. Mengatur Folder Download Manga

Tombol dengan ikon **PlaylistAddFilled** di sisi kanan baris kedua membuka dialog:

1. Dialog berjudul **Set Config Download Manga Directory**.
2. Di dalamnya ada satu kolom teks:
   - Isi dengan path folder di komputer Anda, misalnya:
     - `D:\Manga`
     - `E:\Download\Manga`
3. Klik **Confirm** untuk menyimpan.

Setelah disimpan:

- Folder ini akan dipakai sebagai lokasi utama penyimpanan manga oleh fitur download di halaman lain.
- Folder ini juga dipakai sebagai base folder untuk **Download (test)** di halaman Settings.
- Anda tetap bisa mengubahnya lagi jika perlu.

---

## 7. Cara Kerja Umum Bagi Pengguna

Berikut alur sederhana yang bisa diikuti pengguna akhir:

1. **Pilih atau buat rule untuk suatu situs**
   - Jika rule sudah pernah dibuat:
     - Klik **Load Rules**, pilih situs yang diinginkan.
   - Jika belum:
     - Minta pembuat rule mengisi tab **Manga Rule** dan **Chapter Rule**.

2. **Pastikan rule aktif dan valid**
   - Pastikan switch **enabled** menyala (enabled).
   - Pastikan ikon **Manga Rule** dan **Chapter Rule** berwarna hijau.

3. **Uji rule dengan URL nyata**
   - Copy URL halaman manga dari browser → paste ke kolom **Manga Rule URL** → klik **GO**.
   - Cek tab **Scrape Result** apakah judul dan daftar chapter muncul.
   - Copy URL halaman chapter → paste ke kolom **Chapter Rule URL** → klik **GO**.
   - Cek tab **Scrape Result** apakah muncul daftar URL gambar (pages).

4. **Simpan rule**
   - Jika semua sudah benar, klik **Save Rules**.
   - Rule kini akan tersedia di halaman **Download** dan fitur lain yang menggunakan scraping.

5. **Atur folder download (opsional tapi disarankan)**
   - Klik tombol pengaturan folder (ikon playlist).
   - Masukkan path folder penyimpanan manga.
   - Klik **Confirm**.

Setelah langkah-langkah ini, Anda bisa berpindah ke halaman **Download**, memilih situs dan memasukkan URL manga untuk mulai mengunduh dengan nyaman.

---

## 8. Tips Penggunaan untuk End User

- Jika Anda bukan pembuat rule:
  - Fokus pada pemilihan rule (Load Rules), pengaturan folder, dan pengecekan hasil di **Scrape Result**.
  - Hindari mengubah isi editor **Manga Rule** dan **Chapter Rule** jika tidak yakin.

- Jika setelah klik GO hasil kosong atau aneh:
  - Pastikan URL yang dimasukkan sesuai:
    - Manga Rule URL → halaman detail manga.
    - Chapter Rule URL → halaman baca chapter.
  - Bila tetap tidak berhasil, kemungkinan situs mengubah strukturnya dan rule perlu diperbarui oleh pembuat rule.

- Jika tombol **Save Rules** tidak aktif:
  - Pastikan:
    - `site_key`, `name`, dan `domains` terisi.
    - Dua editor rule tidak kosong.
    - Ikon Manga Rule dan Chapter Rule sudah hijau.

Dengan memahami bagian-bagian di atas, pengguna akhir dapat:

- Memilih, menguji, dan memakai scraping rule dengan aman
- Tanpa perlu memahami detail teknis seperti schema JSON atau logika internal aplikasi.

---

## 9. Panduan Singkat untuk Pembuat Rule

Bagian ini ditujukan untuk **pembuat rule** (power user / dev) yang ingin menambahkan dukungan situs baru atau memperbaiki rule situs lama langsung dari halaman **Settings**.

Jika Anda hanya pengguna biasa, bagian ini bisa dilewati.

### 9.1. Konsep Dasar

Setiap situs manga yang didukung aplikasi umumnya membutuhkan:

1. **Manga Rule**
   - Untuk mengambil:
     - Judul manga
     - Cover
     - Daftar chapter (beserta id, nomor, grup, bahasa, waktu, dll)

2. **Chapter Rule**
   - Untuk mengambil:
     - Daftar URL gambar per halaman dalam satu chapter.

Kedua rule ini ditulis dalam format JSON dan dikerjakan di tab:

- **Manga Rule** (untuk halaman detail manga)
- **Chapter Rule** (untuk halaman baca chapter)

### 9.2. Langkah Umum Membuat Rule Baru

Urutan kerja yang disarankan:

1. **Pilih target situs**
   - Buka situs manga di browser.
   - Pilih satu halaman detail manga dan satu halaman chapter sebagai contoh.

2. **Isi Manga Rule**
   - Masuk ke tab **Manga Rule**.
   - Gunakan snippet/template (jika tersedia di editor) atau copy dari rule lain sebagai acuan.
   - Sesuaikan bagian berikut:
     - `site` → nama situs (misal `"Asmotoon"`).
     - `domains` → daftar domain situs (misal `["asmotoon.com"]`).
     - `strategy` → cara scraping (`"static"`, `"browser"`, `"api"`, atau `"auto"`).
     - `entry.url` → contoh URL halaman detail manga.
     - Bagian `extract` untuk:
       - `title` → ambil judul manga.
       - `cover` → ambil URL cover.
       - `chapters` → ambil daftar chapter.

3. **Isi Chapter Rule**
   - Masuk ke tab **Chapter Rule**.
   - Mulai dari template:
     - Pastikan `site`, `domains`, dan `strategy` konsisten dengan Manga Rule.
     - `entry.url` → contoh URL halaman chapter.
   - Di bagian `extract`, buat field:
     - `pages` → mengambil semua URL gambar halaman.

4. **Tes Manga Rule**
   - Copy URL halaman manga ⇒ paste ke **Manga Rule URL** ⇒ klik **GO**.
   - Periksa tab **Scrape Result**:
     - Pastikan judul dan daftar chapter muncul.
     - Cek apakah `chapter_id` terlihat benar (misal berupa URL atau ID).

5. **Tes Chapter Rule**
   - Dari hasil manga, ambil salah satu `chapter_id` (biasanya berupa URL).
   - Paste ke **Chapter Rule URL** ⇒ klik **GO**.
   - Periksa tab **Scrape Result**:
     - Pastikan ada field `pages` yang berisi daftar URL gambar.
     - Jika tombol **Download** aktif, berarti struktur `pages` sudah valid.

6. **Aktifkan dan simpan**
   - Pastikan switch **enabled** menyala.
   - Pastikan ikon **Manga Rule** dan **Chapter Rule** sudah hijau.
   - Klik **Save Rules** untuk menyimpan.

### 9.3. Tips Mengisi Field Penting Manga Rule

Secara garis besar, Manga Rule membutuhkan:

- `title`
  - Ambil dari elemen judul manga di halaman (biasanya `<h1>` atau sejenis).
  - Contoh pendekatan:
    - `type: "css"`, `selector: "h1.manga-title"`, `trim: true`.

- `cover`
  - Ambil dari `<img>` cover utama.
  - Bisa berupa:
    - `selector: ".cover img"`, `attr: ["src"]`.

- `chapters`
  - Harus berupa list (multiple: true) dan memiliki anak:
    - `chapter_id` → biasanya `href` link chapter atau ID unik.
    - `chapter` → nomor chapter (teks).
    - `group_name` → dimasukkan teks statis atau diambil dari halaman (jika ada).
    - `language` → bisa diisi teks statis `"id"` atau `"en"` jika tidak ada info di halaman.
    - `time` → tanggal/waktu rilis (jika ada).

Contoh pola umum:

- Gunakan selector pada elemen list chapter (misal `.chapter-item` atau `.chapters li`).
- Di setiap item, ambil:
  - `href` link chapter untuk `chapter_id`.
  - Text nomor atau judul chapter untuk `chapter`.
  - Elemen tanggal untuk `time`, jika tersedia.

### 9.4. Tips Mengisi Field `pages` di Chapter Rule

Tujuan field `pages` adalah menghasilkan list URL gambar yang berurutan:

- Jika gambar langsung ada di HTML:
  - Gunakan `type: "css"`, `selector` ke elemen `<img>` halaman.
  - Ambil `attr: ["src"]`.
  - Set `multiple: true`.

- Jika gambar di-load melalui API:
  - Gunakan `strategy: "api"` atau `"auto"` dengan bagian `api.steps`.
  - Pastikan field `pages` mengambil dari hasil API (melalui `from` dan `path` jika menggunakan JSON).

Yang penting untuk aplikasi:

- Hasil akhirnya harus berupa:
  - `pages: ["https://.../page1.jpg", "https://.../page2.jpg", ...]`.

### 9.5. Validasi Visual di Settings

Sebagai pembuat rule, Anda bisa memanfaatkan indikator berikut:

- Ikon **Manga Rule**:
  - Hijau → field kunci (`title`, `cover`, `chapters`) dianggap tersedia.
  - Merah → salah satu belum terdeteksi, cek kembali bagian `extract`.

- Ikon **Chapter Rule**:
  - Hijau → field `pages` terdeteksi.
  - Merah → cek kembali definisi field `pages`.

- Tombol **Download**:
  - Aktif → hasil `Scrape Result` berisi `pages` yang lolos pengecekan format URL.
  - Nonaktif → `pages` kosong / tidak valid.

### 9.6. Workflow Pengembangan Rule

Untuk pengembangan atau debugging rule:

1. Mulai dari satu situs dulu (satu Manga Rule + satu Chapter Rule).
2. Fokus membuat Manga Rule sampai:
   - Title benar.
   - Daftar chapter keluar lengkap.
3. Lanjutkan membuat Chapter Rule:
   - Pastikan semua gambar halaman terbaca urut.
4. Setelah stabil:
   - Simpan rule.
   - Uji di halaman **Download** dengan URL yang sama.

Jika di kemudian hari situs berubah:

- Kembali ke Settings → **Load Rules** → pilih situs yang ingin diperbaiki.
- Update selector / struktur rule.
- Uji lagi dengan URL contoh yang sama.
- Simpan ulang rule setelah perbaikan.

---

### 9.7. Fitur Lanjutan Strategy `"browser"`: `browser.steps` + `actions`

Beberapa situs (terutama yang heavy JS) tidak bisa di-scrape hanya dengan membuka 1 URL lalu mengambil HTML, karena:

- Data gambar/chapters baru muncul setelah halaman lain dibuka terlebih dulu (stateful flow).
- Halaman reader baru merender setelah ada interaksi pengguna (click pada elemen tertentu).

Untuk kasus seperti ini, gunakan fitur `browser.steps` agar dalam **1 kali scrape** aplikasi bisa:

1. Membuka halaman A (mis. detail)
2. Menjalankan aksi (mis. click pada item chapter)
3. Berpindah ke halaman B (mis. reader)
4. Baru melakukan `extract` pada halaman terakhir

#### 9.7.1. Struktur Umum

Tambahkan object `browser` di root rule:

```json
{
  "strategy": "browser",
  "browser": {
    "steps": [
      {
        "url": "https://example.com/detail?slug={id}",
        "wait_config": {
          "content_selectors": [".chapter-item"]
        },
        "actions": [
          { "type": "click", "selector": ".chapter-item[data-num='{chapter_num}']" }
        ]
      },
      {
        "url": "https://example.com/reader?slug={chapter_slug}",
        "wait_config": {
          "content_selectors": ["#reader img"]
        }
      }
    ]
  }
}
```

Keterangan:

- `browser.steps[]` dieksekusi berurutan pada **tab yang sama**.
- `steps[].url` bisa berisi placeholder `{...}` yang akan diganti dari context.
- `steps[].wait_config` opsional: digunakan untuk menunggu halaman step tersebut “siap” sebelum lanjut.
- `steps[].actions[]` opsional: menjalankan interaksi (click / eval) di halaman step tersebut.

#### 9.7.2. Placeholder `{...}` dan Context yang Tersedia

Placeholder akan diganti dari context yang dibangun dari:

- `id` → hasil ekstraksi dari URL input (override URL) berdasarkan `entry.url` / `entry.regex`
- Query param dari URL input → otomatis masuk context  
  Contoh: jika Chapter Rule URL adalah `...?chapter_num=50.1&manga_slug=foo`  
  maka context berisi `chapter_num = "50.1"` dan `manga_slug = "foo"`

Praktiknya, ini memungkinkan pembuat rule “mengirim parameter tambahan” lewat `chapter_id` di Manga Rule, lalu dipakai oleh Chapter Rule.

#### 9.7.3. `actions`: Jenis yang Didukung

- `click`
  - Klik elemen berdasarkan CSS selector.
  - Cocok untuk kasus “harus user click” agar event listener di web berjalan.
  - Properti:
    - `selector` (wajib)
    - `wait_config` (opsional) → dipakai setelah click jika perlu menunggu DOM berubah / pindah halaman

- `eval`
  - Menjalankan JavaScript di halaman.
  - Cocok untuk:
    - Dispatch click custom (mis. `dispatchEvent`)
    - Scroll untuk memicu lazy-load
    - Mengisi localStorage/sessionStorage jika dibutuhkan oleh situs
  - Properti:
    - `script` (wajib) → JS function string yang dieksekusi di browser (`page.Eval`)
    - `wait_config` (opsional) → dipakai setelah script dieksekusi

#### 9.7.4. Urutan `wait_config` yang Dipakai

Untuk strategy `"browser"`, aturan menunggu bisa ditaruh di beberapa tempat:

- `wait_config` di root rule (default)
- `browser.steps[].wait_config` (override untuk step tertentu)
- `browser.steps[].actions[].wait_config` (override setelah action tertentu)

Urutan prioritasnya: `actions[].wait_config` → `steps[].wait_config` → `rule.wait_config`.

#### 9.7.5. Contoh Kasus Nyata: Detail → Click Chapter Item → Reader (stateful)

Untuk situs yang menyimpan data gambar di halaman detail lalu merendernya di reader:

1. **Manga Rule** menghasilkan `chapter_id` yang membawa parameter tambahan, misalnya:
   - `https://mikoroku.com/detail?slug=imaizumin-deep&chapter_num=50.1`
2. **Chapter Rule**:
   - Step 1 membuka detail pakai `{id}` (slug manga)
   - Step 1 click item chapter berdasarkan `{chapter_num}`
   - Tunggu sampai `#reader img` muncul
   - Extract `pages` dari `#reader img`

Contoh Chapter Rule:

```json
{
  "site": "mikoroku",
  "domains": ["mikoroku.com"],
  "strategy": "browser",
  "entry": { "url": "https://mikoroku.com/detail?slug={id}" },
  "browser": {
    "steps": [
      {
        "url": "https://mikoroku.com/detail?slug={id}",
        "wait_config": {
          "content_selectors": [".chapter-item"],
          "timeout_ms": 15000,
          "poll_ms": 150
        },
        "actions": [
          {
            "type": "click",
            "selector": ".chapter-item[data-num='{chapter_num}']",
            "wait_config": {
              "content_selectors": ["#reader img"],
              "timeout_ms": 15000,
              "poll_ms": 150,
              "skip_render_stable": true
            }
          }
        ]
      }
    ]
  },
  "extract": [
    {
      "name": "pages",
      "type": "css",
      "selector": "#reader img",
      "multiple": true,
      "attr": ["data-original", "data-src", "src"]
    }
  ]
}
```

Jika situs tidak punya atribut `data-num`, alternatifnya gunakan `data-index` (mis. `.chapter-item[data-index='{chapter_index}']`) dan isi `chapter_index` lewat query param yang disimpan di `chapter_id`.

## 10. Penjelasan Lengkap Fitur `extract` dan Tipe `type`

Bagian ini fokus ke cara konfigurasi **field di dalam `extract`**, karena di sinilah Anda mengatur:

- Apa yang diambil dari halaman
- Dari mana sumber datanya
- Bagaimana hasilnya dibentuk (satu nilai atau list, plain text atau hasil regex, dll)

### 10.1. Struktur Umum Satu Field `extract`

Secara umum, satu item di dalam array `extract` berbentuk seperti ini:

```json
{
  "name": "title",
  "type": "css",
  "selector": "h1.manga-title",
  "multiple": false,
  "trim": true
}
```

Field umum yang sering dipakai:

- `name` (wajib)
  - Nama field di hasil akhir.
  - Contoh: `"title"`, `"cover"`, `"chapters"`, `"pages"`, `"group_name"`, dll.

- `type` (wajib)
  - Menentukan **cara** mengambil data.
  - Nilai yang diperbolehkan:
    - `"css"`
    - `"json"`
    - `"template"`
    - `"text"`

- `selector`
  - Dipakai oleh `type: "css"`.
  - Berisi string selector CSS, contoh:
    - `"h1"`
    - `".chapter-item a"`
    - `"#content .title"`

- `attr`
  - Array berisi nama atribut HTML yang ingin diambil.
  - Umumnya dipakai untuk mengambil `href` atau `src`.
  - Contoh:
    - `"attr": ["href"]`
    - `"attr": ["src"]`

- `multiple`
  - `true` → hasil berupa list/array.
  - `false` atau tidak diisi → hasil satu nilai.
  - Untuk field seperti `chapters` dan `pages`, biasanya `multiple: true`.

- `trim`
  - Jika `true`, hasil teks akan di-trim dari spasi berlebih di awal/akhir.

- `regex`
  - Dipakai untuk memproses teks yang sudah diambil.
  - Contoh:
    - Mengambil ID dari URL: `"/chapter/([^/]+)"`.

- `replace`
  - Dipakai untuk mengganti teks menggunakan **regex replace** setelah nilai berhasil diambil.
  - Berguna untuk:
    - Menghapus prefix/suffix (mis. `"Chapter "`).
    - Membersihkan whitespace berlebih.
    - Menghapus query string yang tidak perlu, dsb.

- `children`
  - Dipakai ketika field `multiple` berisi objek kompleks.
  - Contoh:
    - Field `chapters` yang di dalamnya punya `chapter_id`, `chapter`, `group_name`, dll.

### 10.2. Tipe `type: "css"`

`type: "css"` adalah tipe yang paling umum dipakai.

Kegunaan:

- Mengambil data dari HTML menggunakan selector CSS:
  - Text di dalam elemen
  - Nilai atribut (href, src, title, dll)

Pola dasar:

```json
{
  "name": "title",
  "type": "css",
  "selector": "h1.manga-title",
  "trim": true
}
```

Artinya:

- Cari elemen yang cocok dengan `h1.manga-title`.
- Ambil teksnya.
- Trim spasi di kiri kanan.

Contoh lain untuk mengambil link:

```json
{
  "name": "chapter_id",
  "type": "css",
  "selector": ".chapter-item a",
  "attr": ["href"]
}
```

Jika dipakai di dalam field `chapters` yang `multiple: true`:

- Untuk setiap elemen `.chapter-item`, ambil `<a>` di dalamnya.
- Dari `<a>`, ambil atribut `href`.

Dengan `regex`, Anda bisa memotong URL:

```json
{
  "name": "chapter_id",
  "type": "css",
  "selector": ".chapter-item a",
  "attr": ["href"],
  "regex": "/chapter/([^/]+)"
}
```

Ini berguna kalau Anda hanya ingin mengambil ID di tengah URL.

Dengan `replace`, Anda bisa membersihkan hasil akhir menggunakan regex replace.

Contoh: buang prefix "Chapter " (case-insensitive) dari judul chapter:

```json
{
  "name": "chapter",
  "type": "css",
  "selector": ".chapter-title",
  "trim": true,
  "replace": {
    "pattern": "(?i)^chapter\\s+",
    "with": ""
  }
}
```

Contoh: bersihkan whitespace berlebih menjadi satu spasi:

```json
{
  "name": "title",
  "type": "css",
  "selector": "h1",
  "trim": true,
  "replace": {
    "pattern": "\\s+",
    "with": " "
  }
}
```

Urutan pemrosesan yang dipakai aplikasi:

- Untuk `type: "css"`:
  - Ambil `attr` (jika ada) atau text elemen → `trim` → `regex` → `replace`
- Untuk `type: "json"`:
  - Ambil string dari `path` → `regex` → `replace`

### 10.3. Tipe `type: "json"`

Dipakai ketika sumber data berupa **JSON**, biasanya dari API (digunakan bersama `strategy: "api"` atau `"auto"`).

Untuk membaca JSON, aplikasi menggunakan **GJSON** (library Go).  
Artinya, nilai `path` di sini mengikuti **GJSON path syntax**.

Field tambahan penting:

- `path` (wajib jika `type: "json"`)
  - Menentukan lokasi data di dalam struktur JSON dengan gaya GJSON.

#### 10.3.1. Bentuk Path Dasar (GJSON)

Beberapa pola dasar yang sering dipakai:

- Akses field biasa:
  - `data.title` → membaca `{ "data": { "title": "..." } }`
  - `manga.id` → membaca `{ "manga": { "id": "..." } }`

- Akses array dengan index:
  - `chapters.0.id` → element pertama array `chapters` lalu field `id`
  - `chapters.1.language` → element kedua array `chapters` lalu field `language`

- Mengambil panjang array:
  - `chapters.#` → jumlah elemen di `chapters`
  - `pages.#` → jumlah halaman di `pages`

- Mengambil seluruh array:
  - `chapters` → akan mengembalikan seluruh array `chapters`
  - `pages` → seluruh list URL di `pages`

- Field dengan titik di dalam nama (harus di-escape):
  - `meta\.data.value` → membaca field dengan nama `meta.data`

#### 10.3.2. Query di Dalam Array

GJSON juga mendukung query di dalam array, misalnya:

```json
{
  "chapters": [
    { "id": "c1", "language": "id", "group": "A" },
    { "id": "c2", "language": "en", "group": "B" },
    { "id": "c3", "language": "id", "group": "B" }
  ]
}
```

Contoh path yang valid:

- Ambil semua `id` chapter dengan bahasa Indonesia:
  - `chapters.#(language=="id")#.id`
- Ambil semua `id` chapter dengan group `B`:
  - `chapters.#(group=="B")#.id`
- Ambil `id` pertama yang bahasa Inggris:
  - `chapters.#(language=="en").id`

Operator yang bisa dipakai di dalam `#(...)` antara lain:

- `==`, `!=`, `<`, `<=`, `>`, `>=`
- `%` (like) dan `!%` (not like) untuk pencocokan sederhana.

#### 10.3.3. Contoh Path Praktis untuk Scraping Rule

Misalkan response API seperti ini:

```json
{
  "data": {
    "title": "My Manga",
    "cover": "https://img.example.com/cover.jpg",
    "chapters": [
      { "id": "c1", "chapter": "1", "language": "id" },
      { "id": "c2", "chapter": "2", "language": "en" }
    ]
  }
}
```

Contoh path yang bisa dipakai di `path`:

- Ambil judul:
  - `data.title`
- Ambil URL cover:
  - `data.cover`
- Ambil seluruh array chapters:
  - `data.chapters`
- Ambil `id` chapter pertama:
  - `data.chapters.0.id`
- Ambil semua `id` chapter bahasa Indonesia:
  - `data.chapters.#(language=="id")#.id`

Untuk API khusus chapter pages, contoh JSON:

```json
{
  "data": {
    "pages": [
      "https://img.example.com/manga/1/01.jpg",
      "https://img.example.com/manga/1/02.jpg"
    ]
  }
}
```

Maka field `pages` di rule bisa:

- `path: "data.pages"` → langsung ambil array `pages`.

#### 10.3.4. Contoh Konfigurasi Field

Contoh sederhana:

```json
{
  "name": "title",
  "type": "json",
  "from": "step1",
  "path": "data.title"
}
```

Artinya:

- Ambil response dari API step dengan `id: "step1"`.
- Dari JSON tersebut, baca `data.title` memakai path GJSON.

Jika data berupa list:

```json
{
  "name": "pages",
  "type": "json",
  "from": "step1",
  "path": "data.pages",
  "multiple": true
}
```

Artinya:

- Ambil `data.pages` dari JSON.
- Asumsinya sudah berupa array string (URL gambar).

### 10.4. Tipe `type: "template"`

Dipakai untuk membentuk string baru dari beberapa field lain atau dari nilai yang sudah diambil.

Field tambahan penting:

- `template` (wajib jika `type: "template"`)

Polanya biasanya seperti:

```json
{
  "name": "full_title",
  "type": "template",
  "template": "{{title}} - {{language}}"
}
```

Contoh kegunaan:

- Menggabungkan judul dengan bahasa, chapter number dengan judul, dsb.

Catatan:

- Detil penggantian placeholder tergantung implementasi di backend, tetapi konsepnya Anda menyusun string dari nama field yang sudah ada.

### 10.5. Tipe `type: "text"`

Dipakai ketika Anda ingin mengisi field dengan teks **statis**.

Field tambahan penting:

- `text` (wajib jika `type: "text"`)

Contoh:

```json
{
  "name": "group_name",
  "type": "text",
  "text": "Default Fansub"
}
```

Atau untuk bahasa:

```json
{
  "name": "language",
  "type": "text",
  "text": "id"
}
```

Kegunaan:

- Mengisi nilai yang tidak berubah-ubah antar item, misalnya:
  - Nama grup kalau situs tidak menampilkan grup per chapter.
  - Bahasa default situs.

### 10.6. Field `multiple` dan `children`

`multiple` dan `children` sangat penting untuk field yang menghasilkan daftar objek, seperti:

- `chapters` di Manga Rule
- `pages` (jika ingin struktur kompleks di masa depan)

Contoh pola `chapters`:

```json
{
  "name": "chapters",
  "type": "css",
  "selector": ".chapter-item",
  "multiple": true,
  "children": [
    {
      "name": "chapter_id",
      "type": "css",
      "selector": "a",
      "attr": ["href"]
    },
    {
      "name": "chapter",
      "type": "css",
      "selector": ".chapter-number"
    },
    {
      "name": "group_name",
      "type": "text",
      "text": "Default Group"
    },
    {
      "name": "language",
      "type": "text",
      "text": "id"
    },
    {
      "name": "time",
      "type": "css",
      "selector": ".chapter-date"
    }
  ]
}
```

Cara bacanya:

- `selector: ".chapter-item"` → cari semua elemen chapter di halaman.
- `multiple: true` → tiap elemen akan menjadi satu objek dalam array `chapters`.
- `children`:
  - Menjelaskan bagaimana mengisi properti di dalam setiap chapter.

Aplikasi mengharapkan:

- Untuk Manga Rule, field `chapters` memiliki minimal child:
  - `chapter_id`, `chapter`, `group_name`, `language`, `time`.

### 10.7. Field `from` untuk Data API

`from` digunakan untuk menunjukkan sumber data ketika Anda memakai beberapa **api step**.

Contoh konfigurasi API:

```json
"api": {
  "steps": [
    {
      "id": "step1",
      "request": {
        "url": "https://api.example.com/manga/{id}"
      },
      "response": "json"
    }
  ]
}
```

Kemudian di `extract`:

```json
{
  "name": "title",
  "type": "json",
  "from": "step1",
  "path": "data.title"
}
```

Artinya:

- Ambil data dari response step1.
- Baca field `data.title` di JSON.

`from` juga bisa dipakai oleh tipe `template` jika nantinya mendukung referensi field yang berasal dari API.

### 10.8. Filter dan `filter_mode`

`filter` dan `filter_mode` bisa digunakan untuk menyaring elemen yang tidak diinginkan.

Contoh:

```json
{
  "name": "chapters",
  "type": "css",
  "selector": ".chapter-item",
  "multiple": true,
  "filter": "Bonus",
  "filter_mode": "not",
  "children": [
    {
      "name": "chapter",
      "type": "css",
      "selector": ".chapter-number"
    }
  ]
}
```

Artinya:

- Ambil semua `.chapter-item`.
- Singkirkan yang mengandung teks `"Bonus"` (karena `filter_mode: "not"`).

Atau:

```json
{
  "filter": "Chapter",
  "filter_mode": "has"
}
```

→ hanya ambil elemen yang mengandung kata `"Chapter"`.

### 10.9. Strategi Memilih `type` yang Tepat

Ringkasan kapan memakai tipe apa:

- Pakai **`css`** ketika:
  - Data ada di HTML halaman.
  - Anda bisa mengaksesnya dengan selector CSS standar.

- Pakai **`json`** ketika:
  - Data diambil dari endpoint API JSON.
  - Anda sudah definisikan `api.steps` dan ingin mengambil bagian tertentu dari response.

- Pakai **`template`** ketika:
  - Anda ingin menggabungkan beberapa nilai menjadi satu string.
  - Contoh: menggabungkan judul dan nomor chapter.

- Pakai **`text`** ketika:
  - Nilai bersifat tetap / default.
  - Contoh: bahasa `"id"`, nama grup default, dsb.

Dengan memahami kombinasi:

- `type`
- `selector` / `path`
- `multiple`
- `children`
- `attr`, `regex`, `filter`, `from`

Anda bisa membangun rule yang fleksibel dan cukup kuat untuk berbagai struktur situs manga berbeda.
