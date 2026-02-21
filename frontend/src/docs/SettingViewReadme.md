<!-- markdownlint-disable-all -->

# SettingsView – Dokumentasi Scraping Rule

File terkait:

- [SettingsView.vue](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/views/SettingsView.vue)
- [MangaRuleSchema.json](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/assets/MangaRuleSchema.json)
- [ChapterRuleSchema.json](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/assets/ChapterRuleSchema.json)
- [validationHelpers.ts](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/utils/validationHelpers.ts)
- [scrape.ts (tipe hasil scrape)](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/type/scrape.ts)

Dokumentasi ini menjelaskan:

- Cara memakai halaman **Settings** untuk membuat dan menguji scraping rule
- Struktur lengkap JSON **Manga Rule** dan **Chapter Rule**
- Validasi yang dilakukan aplikasi
- Hubungan scraping rule dengan fitur **Download** dan direktori manga

---

## 1. Gambaran Umum SettingsView

Halaman Settings menyediakan 3 area utama:

- Baris atas (toolbar):
  - Input **Manga Rule URL** + tombol **GO** → test Manga Rule ke satu URL
  - Input **Chapter Rule URL** + tombol **GO** → test Chapter Rule ke satu URL
  - Tombol **Download** → test download berdasarkan hasil JSON di panel `Scrape Result`
  - Tombol **Save Rules** → simpan rule ke database
  - Tombol **Load Rules** → buka modal untuk memilih rule yang sudah disimpan

- Baris kedua (metadata rule):
  - `site_key` (otomatis diisi dari field `site` di JSON)
  - `name` (otomatis diisi dari field `site` di JSON)
  - `domains` (otomatis diisi dari field `domains` di JSON)
  - Switch **enabled/disabled**
  - Indikator status JSON:
    - **Manga Rule** (ikon hijau/merah)
    - **Chapter Rule** (ikon hijau/merah)

- Baris ketiga (editor):
  - Tab **Manga Rule** → editor JSON dengan schema `MangaRuleSchema`
  - Tab **Chapter Rule** → editor JSON dengan schema `ChapterRuleSchema`
  - Tab **Scrape Result** → menampilkan JSON hasil `ScraperService.Scrape`

Tambahan:

- Tombol **Set Config Download Manga Directory** untuk mengatur direktori utama download manga (`manga_directory`).
- Modal **Load Scraping Rules** yang menampilkan daftar rule dari database dan mengisi form saat dipilih.

---

## 2. Alur Kerja Dasar

1. Tulis / modifikasi JSON **Manga Rule** dan **Chapter Rule** di editor.
2. Pastikan kedua JSON lolos validasi (ikon Manga Rule dan Chapter Rule berwarna hijau).
3. Isi URL uji:
   - Masukkan URL manga ke **Manga Rule URL** lalu klik **GO** → menguji Manga Rule.
   - Masukkan URL chapter ke **Chapter Rule URL** lalu klik **GO** → menguji Chapter Rule.
4. Lihat hasil di tab **Scrape Result**:
   - Untuk Manga Rule, seharusnya berbentuk `MangaData`.
   - Untuk Chapter Rule, seharusnya berbentuk `ChapterPages`.
5. Jika hasil sesuai dan validasi hijau:
   - Klik **Save Rules** untuk menyimpan ke database.
6. Rule yang tersimpan akan muncul di:
   - Modal **Load Rules** di SettingsView.
   - Dropdown pemilihan situs di **DownloadView**.

---

## 3. Struktur Data Hasil Scrape

Tipe hasil scrape didefinisikan di [scrape.ts](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/type/scrape.ts).

### 3.1. MangaData (hasil dari Manga Rule)

```ts
export interface MangaData {
  id: string
  title: string
  cover: string
  chapters: ChapterData[]
  total_pages?: number
}

export interface ChapterData {
  chapter_id: string
  chapter: number | string
  chapter_title?: string
  chapter_volume?: string
  group_name: string
  language: string
  time: string
  status?: boolean
}
```

Interpretasi:

- `title` → judul manga
- `cover` → URL cover utama
- `chapters` → daftar chapter yang akan muncul di DownloadView
  - `chapter_id` → biasanya URL chapter atau ID unik lain
  - `chapter` → nomor chapter (string atau number)
  - `group_name` → nama kelompok penerjemah
  - `language` → bahasa (misal `id`, `en`)
  - `time` → informasi waktu (misal tanggal rilis)

### 3.2. ChapterPages (hasil dari Chapter Rule)

```ts
export interface ChapterPages {
  pages: string[]
}
```

`pages` berisi daftar URL gambar halaman dalam satu chapter. Tipe ini digunakan:

- Di SettingsView untuk test tombol **Download** (cek validitas via `isValidPages`)
- Di DownloadView untuk proses download sebenarnya.

---

## 4. Struktur Umum JSON Scraping Rule

Baik Manga Rule maupun Chapter Rule memakai struktur dasar yang sama (lihat `MangaRuleSchema.json` dan `ChapterRuleSchema.json`):

```json
{
  "site": "Nama Situs",
  "domains": ["example.com"],
  "strategy": "static",
  "entry": {
    "url": "https://example.com/manga/123",
    "method": "GET",
    "regex": "",
    "headers": {
      "User-Agent": "..."
    }
  },
  "wait_config": {
    "container_selectors": ["#main"],
    "content_selectors": [".chapter-list"],
    "min_text_length": 100,
    "require_image_loaded": true,
    "timeout_ms": 10000,
    "poll_ms": 200,
    "skip_waits": false,
    "skip_render_stable": false,
    "skip_navigation_wait": false,
    "navigation_timeout_ms": 15000
  },
  "api": {
    "steps": [
      {
        "id": "step1",
        "request": {
          "url": "https://api.example.com/manga/{id}",
          "method": "GET",
          "headers": {
            "Authorization": "Bearer ..."
          }
        },
        "response": "json"
      }
    ]
  },
  "extract": [
    {
      "name": "title",
      "type": "css",
      "selector": "h1"
    }
  ]
}
```

### 4.1. Field Wajib (top-level)

- `site` (string, required)
  - Nama situs, juga digunakan untuk:
    - Mengisi otomatis field `name` di form
    - Membentuk `site_key` (lowercase + tanpa spasi)

- `domains` (array string, required)
  - Daftar domain yang digunakan untuk:
    - Menentukan rule mana yang dipakai saat user paste URL di DownloadView.
  - Contoh: `["mangadomain.com", "www.mangadomain.com"]`

- `strategy` (string, required)
  - Nilai: `"static" | "browser" | "api" | "auto"`
  - Mengatur cara Scraper memuat halaman:
    - `static` → request HTTP biasa, cocok untuk halaman statis.
    - `browser` → menggunakan browser/headless, untuk situs heavy JS.
    - `api` → hanya lewat endpoint API (lihat bagian `api`).
    - `auto` → bisa menggunakan `entry` atau `api`, tergantung konfigurasi.

- `extract` (array, required)
  - Daftar aturan field yang akan diekstrak dari HTML/JSON/API.
  - Bentuk setiap item diatur oleh definisi `fieldRule`.

### 4.2. `entry` (opsional tapi required untuk beberapa strategy)

Dipakai jika `strategy` adalah `static` atau `browser`. Schema:

- `url` (string, format `uri`, required)
- `method` (string, opsional, default biasanya `GET`)
- `regex` (string, opsional)
  - Bisa dipakai untuk mengekstrak ID dari URL awal.
- `headers` (object string→string, opsional)
  - Override header HTTP (User-Agent, cookie, dll).

Aturan schema:

- Jika `strategy = "api"` → `api` wajib, `entry` opsional.
- Jika `strategy = "static"` atau `"browser"` → `entry` wajib.
- Jika `strategy = "auto"` → wajib punya minimal salah satu: `api` atau `entry`.

### 4.3. `wait_config` (opsional, untuk strategy `browser`/`static`)

Digunakan untuk mengontrol kapan scraper menganggap halaman sudah siap di-scrape:

- `container_selectors` (array string)
  - Selektor CSS container utama konten.
- `content_selectors` (array string)
  - Selektor elemen yang menandakan konten sudah muncul.
- `min_text_length` (integer)
  - Minimal panjang teks sebelum dianggap siap.
- `require_image_loaded` (boolean)
  - Jika `true`, menunggu sampai gambar termuat.
- `timeout_ms`, `navigation_timeout_ms` (integer)
  - Batas waktu wait.
- `poll_ms` (integer)
  - Interval pengecekan kondisi.
- `skip_waits`, `skip_render_stable`, `skip_navigation_wait` (boolean)
  - Mengabaikan bagian wait tertentu jika `true`.

### 4.4. `api` (opsional, untuk strategy `api`/`auto`)

Struktur:

```json
"api": {
  "steps": [
    {
      "id": "step1",
      "request": {
        "url": "https://api.example.com/...",
        "method": "GET",
        "headers": {
          "X-Token": "..."
        }
      },
      "response": "json"
    }
  ]
}
```

- `steps` (array, required)
  - Urutan panggilan API.
- Setiap `apiStep`:
  - `id` (string, pattern `^[a-zA-Z_][a-zA-Z0-9_]*$`)
    - Dipakai oleh field `from` di `extract` untuk mengambil data dari step tertentu.
  - `request`:
    - `url` (string, required)
    - `method` (string, optional)
    - `headers` (object string→string, optional)
  - `response` (string, `"json" | "html"`, default `"json"`)

---

## 5. Definisi Field Rule (`extract`)

Setiap item dalam `extract` mengikuti definisi `fieldRule`:

Field umum:

- `name` (string, required)
  - Nama field yang akan muncul di hasil (misal `title`, `cover`, `chapters`, `pages`).
- `type` (string, required)
  - Nilai: `"css" | "json" | "template" | "text"`.
- `from` (string, opsional)
  - Jika diisi, menunjuk ke `id` step API yang akan dipakai sumber datanya.
- `selector` (string, opsional)
  - Selektor CSS untuk mengambil elemen (jika `type = "css"`).
- `attr` (array string, opsional)
  - Daftar atribut yang diambil dari elemen (misal `["href"]`).
- `filter` (string, opsional) dan `filter_mode` (`"has" | "not"`, default `"has"`)
  - Menyaring elemen yang mengandung / tidak mengandung teks tertentu.
- `path` (string, opsional)
  - Path JSON (jika `type = "json"`).
- `multiple` (boolean, default `false`)
  - Jika `true`, hasil berupa array.
- `trim` (boolean, default `false`)
  - Jika `true`, akan di-trim spasi.
- `regex` (string, opsional)
  - Regex untuk memproses teks hasil (misal ekstrak ID dari URL).
- `template` (string, opsional)
  - Template string (jika `type = "template"`).
- `text` (string, opsional, required jika `type = "text"`)
  - Nilai statis yang langsung menjadi hasil field.
- `children` (array fieldRule, opsional)
  - Digunakan untuk field kompleks (misalnya `chapters` sebagai array objek).

Khusus:

- Jika `type = "json"` → `path` wajib.
- Jika `type = "template"` → `template` wajib.
- Jika `type = "text"` → `text` wajib.

---

## 6. Aturan Khusus Manga Rule

Selain schema JSON, Manga Rule juga divalidasi di frontend:

- Fungsi: `validateMangaRule` di [validationHelpers.ts](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/utils/validationHelpers.ts)
- Validasi tambahan:
  - Field `extract` harus mengandung **minimal** rule dengan `name`:
    - `"title"`
    - `"cover"`
    - `"chapters"`
- Jika salah satu hilang → editor akan memberi error, dan status **Manga Rule** di header akan merah (`statusJson.manga_rule = false`).

### 6.1. Field `chapters` di Manga Rule

Schema menetapkan aturan khusus saat `name = "chapters"`:

- `multiple` wajib `true`.
- `children` wajib ada dan minimal 1 item.
- `children` harus mengandung field dengan `name`:
  - `"chapter_id"`
  - `"chapter"`
  - `"group_name"`
  - `"language"`
  - `"time"`

Contoh sederhana:

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

---

## 7. Aturan Khusus Chapter Rule

Chapter Rule memakai schema yang hampir sama dengan Manga Rule, tetapi:

- Validasi tambahan di frontend:
  - Fungsi `validateChapterRule` memastikan `extract` memiliki field dengan `name = "pages"`.
- Struktur hasil:
  - Harus memenuhi tipe `ChapterPages` → `{ "pages": ["https://.../1.jpg", ...] }`.
- Di SettingsView:
  - Hasil JSON disimpan ke `resultJson`.
  - `watchDebounced` memanggil `isValidPages`:
    - Memastikan `pages` ada, berupa array string, dan tiap string adalah URL valid.
    - Jika valid → tombol **Download** aktif (`statusDownload = true`).

Contoh rule `pages` sederhana:

```json
{
  "name": "pages",
  "type": "css",
  "selector": ".reader img",
  "multiple": true,
  "attr": ["src"]
}
```

---

## 8. Otomatisasi Form (`site_key`, `name`, `domains_json`)

Di SettingsView terdapat watcher:

```ts
watchDebounced(
  () => scrapingRuleInput.manga_rule_json,
  m => {
    if (!m) return
    const mm = JSON.parse(m)
    const obj = mm as Record<string, any>
    if ('domains' in obj && Array.isArray(obj.domains)) {
      scrapingRuleInput.domains_json = JSON.stringify(obj.domains)
    }
    if ('site' in obj) {
      scrapingRuleInput.name = obj.site
      scrapingRuleInput.site_key = obj.site.replace(/\s+/g, '').toLowerCase()
    }
  },
  { debounce: 500, maxWait: 1000 },
)
```

Artinya:

- Saat Manga Rule JSON valid dan berubah:
  - Field `domains_json` otomatis diisi dengan `JSON.stringify(domains)`.
  - Field `name` dan `site_key` otomatis diisi dari `site`.
- Dengan demikian:
  - Anda cukup mengubah bagian `site` dan `domains` di Manga Rule.
  - Metadata form akan mengikuti secara otomatis.

---

## 9. Kondisi Siap Simpan (`readyToSave`)

Tombol **Save Rules** hanya aktif jika `readyToSave` bernilai `true`:

```ts
const readyToSave = computed(() => {
  return (
    scrapingRuleInput.site_key &&
    scrapingRuleInput.name &&
    scrapingRuleInput.domains_json &&
    scrapingRuleInput.chapter_rule_json &&
    scrapingRuleInput.manga_rule_json &&
    statusJson.manga_rule &&
    statusJson.chapter_rule
  )
})
```

Syarat:

- `site_key`, `name`, `domains_json` tidak kosong.
- `manga_rule_json` dan `chapter_rule_json` terisi.
- Validasi JSON keduanya **lolos** (ikon hijau).

Saat klik **Save Rules**:

- Data form diubah menjadi instance `ScrapingRule` dan disimpan via `DatabaseService.SaveScrapingRule`.
- Setelah sukses:
  - `clearInput()` dipanggil untuk reset form.
  - `listScrapeRuleDb` di-refresh dari database.

---

## 10. Konfigurasi Direktori Download Manga

SettingsView juga mengelola konfigurasi:

- Key config: `manga_directory`
- Disimpan melalui `DatabaseService.SetConfig('manga_directory', value)`
- Dibaca kembali oleh:
  - SettingsView (untuk menampilkan nilai sekarang)
  - Utilitas path di [filePathHelper.ts](file:///f:/Development/Go/wails3/mangav5-wails3/frontend/src/utils/filePathHelper.ts)
  - Fitur Download (DownloadView dan Reader)

Fungsi penting:

```ts
export const setMangaDirectory = async (directory: string) => {
  await DatabaseService.SetConfig('manga_directory', directory)
}
```

```ts
export async function getDownloadDir(
  title: string = 'untitled',
  chapter: string | number = '000',
): Promise<string> {
  const downloadDir = await getMangaDirectory()
  return `${downloadDir}/${safeWindowsDirectoryName(title)}/${chapter}`
}
```

Rekomendasi:

- Set `manga_directory` ke lokasi yang stabil dan punya ruang cukup.
- Hindari path dengan karakter aneh; fungsi `safeWindowsDirectoryName` sudah membantu sanitasi nama folder berdasarkan judul.

Catatan:

- Tombol **Download** di SettingsView hanya untuk **test cepat** dengan path statis (di kode saat ini masih hardcoded).
- Proses download utama menggunakan konfigurasi `manga_directory` di DownloadView.

---

## 11. Contoh Template Rule

Monaco editor sudah dikonfigurasi dengan snippet bawaan (lihat `scrapingRuleSnippets` di `monacoSnippets.ts`). Anda bisa memanggil snippet dari editor untuk mempercepat penulisan rule.

### 11.1. Contoh Manga Rule (static)

```json
{
  "site": "Asmotoon",
  "domains": ["asmotoon.com"],
  "strategy": "static",
  "entry": {
    "url": "https://asmotoon.com/chapter/6478dd70f1a-6478e2d062d/"
  },
  "extract": [
    {
      "name": "title",
      "type": "css",
      "selector": "h1",
      "trim": true
    },
    {
      "name": "cover",
      "type": "css",
      "selector": ".manga-cover img",
      "attr": ["src"]
    },
    {
      "name": "chapters",
      "multiple": true,
      "type": "css",
      "selector": "#chapters a",
      "children": [
        {
          "name": "chapter_id",
          "type": "css",
          "attr": ["href"],
          "regex": "/chapter/([^/]+)"
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
  ]
}
```

### 11.2. Contoh Chapter Rule (static)

```json
{
  "site": "Asmotoon",
  "domains": ["asmotoon.com"],
  "strategy": "static",
  "entry": {
    "url": "https://asmotoon.com/chapter/6478dd70f1a-6478e2d062d/"
  },
  "extract": [
    {
      "name": "pages",
      "type": "css",
      "selector": ".page img",
      "multiple": true,
      "attr": ["src"]
    }
  ]
}
```

---

## 12. Tips Penggunaan & Debug

- Gunakan selalu **Manga Rule URL** dan **Chapter Rule URL** untuk menguji:
  - Apakah hasil `ScraperService.Scrape` sudah sesuai.
  - Apakah struktur hasil memenuhi `MangaData` / `ChapterPages`.
- Cek tab **Scrape Result**:
  - Jika kosong atau error → cek log di console dan pesan dialog error.
  - Jika struktur tidak sesuai → periksa ulang konfigurasi `extract`.
- Perhatikan indikator status:
  - Ikon Manga Rule / Chapter Rule **harus hijau** sebelum rule dipakai di DownloadView.
- Jika situs berubah struktur HTML/API:
  - Update bagian `selector`, `path`, atau `api.steps` yang relevan.
  - Uji ulang dengan URL yang sama.

Dengan mengikuti dokumentasi ini, Anda bisa:

- Menambahkan scraping rule untuk situs baru.
- Menyesuaikan rule jika struktur situs berubah.
- Memastikan rule valid sebelum dipakai di proses download utama.
