# Panduan Menambahkan Bahasa Baru ke Monaco Editor

Monaco Editor pada proyek ini telah dioptimasi menggunakan `vite-plugin-monaco-editor` dan konfigurasi alias khusus untuk meminimalkan ukuran bundle. Secara default, hanya bahasa **JSON** yang dimuat.

Jika Anda ingin menambahkan dukungan untuk bahasa lain (misalnya HTML, CSS, TypeScript), ikuti 2 langkah berikut:

## 1. Update `vite.config.ts`

Anda perlu mendaftarkan worker bahasa tersebut agar Vite dapat memprosesnya. Tambahkan ID bahasa ke dalam opsi `languageWorkers`.

Buka file `frontend/vite.config.ts`:

```typescript
// ...
monacoEditorPlugin({
  // Tambahkan ID bahasa di sini (contoh: 'html', 'css', 'typescript')
  languageWorkers: ['json', 'html', 'editorWorkerService'], 
  customWorkers: [
    {
      label: 'json',
      entry: 'monaco-editor/esm/vs/language/json/json.worker',
    },
    // (Opsional) Biasanya plugin otomatis mendeteksi, tapi jika error bisa ditambahkan manual:
    // {
    //   label: 'html',
    //   entry: 'monaco-editor/esm/vs/language/html/html.worker',
    // }
  ],
}),
// ...
```

## 2. Update `MonacoEditor.vue`

Karena kita menggunakan alias yang hanya memuat **Core API** (`editor.api`), fitur bahasa tidak dimuat otomatis. Anda harus mengimpor file kontribusi bahasa tersebut secara manual di komponen.

Buka file `frontend/src/components/MonacoEditor.vue`:

```typescript
<script setup lang="ts">
import * as monaco from 'monaco-editor'

// Import bahasa yang sudah ada
import 'monaco-editor/esm/vs/language/json/monaco.contribution'

// === TAMBAHKAN IMPORT DI BAWAH INI ===

// Contoh untuk HTML:
import 'monaco-editor/esm/vs/language/html/monaco.contribution'

// Contoh untuk CSS:
// import 'monaco-editor/esm/vs/language/css/monaco.contribution'

// Contoh untuk TypeScript/JavaScript:
// import 'monaco-editor/esm/vs/language/typescript/monaco.contribution'

// ...
</script>
```

Setelah melakukan kedua langkah di atas, jalankan ulang server dev (`wails3 dev` atau `bun run dev`) agar perubahan config Vite diterapkan.

---

# Mengembalikan Pengaturan ke Semua Bahasa (Default)

Jika Anda ingin membatalkan optimasi dan memuat **semua bahasa** yang didukung Monaco Editor (seperti perilaku default, namun ukuran file akan membesar drastis >5MB), ikuti langkah berikut:

## 1. Hapus Alias di `vite.config.ts`
Hapus konfigurasi alias yang memaksa penggunaan `editor.api.js`.

```typescript
// frontend/vite.config.ts

// Hapus atau komentari bagian ini:
// {
//   find: /^monaco-editor$/,
//   replacement: 'monaco-editor/esm/vs/editor/editor.api.js',
// },
```

## 2. Hapus Filter Bahasa di `vite.config.ts`
Biarkan `monacoEditorPlugin` berjalan tanpa filter `languageWorkers` agar memuat semua worker default.

```typescript
// frontend/vite.config.ts

monacoEditorPlugin({
  // Hapus properti 'languageWorkers' dan 'customWorkers' untuk memuat default (semua bahasa)
}),
```

Setelah langkah ini, `import * as monaco from 'monaco-editor'` di `MonacoEditor.vue` akan secara otomatis memuat seluruh fitur editor beserta semua bahasa yang tersedia.
