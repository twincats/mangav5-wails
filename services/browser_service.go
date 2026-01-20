package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// BrowserService menangani operasi scraping menggunakan browser headless
type BrowserService struct {
	browser *rod.Browser
}

// ScrapeResult menyimpan hasil scraping
type ScrapeResult struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Error   string   `json:"error,omitempty"`
}

// NewBrowserService membuat instance baru dari BrowserService
func NewBrowserService() *BrowserService {
	return &BrowserService{}
}

// initBrowser menginisialisasi browser jika belum ada atau terputus
func (s *BrowserService) initBrowser() error {
	if s.browser != nil {
		// Cek apakah browser masih responsif
		if _, err := s.browser.Version(); err == nil {
			return nil
		}
	}

	// Gunakan launcher untuk mencari browser default dan set headless=true
	// Kita nonaktifkan Leakless untuk menghindari false positive antivirus di Windows
	l := launcher.New().
		Headless(true).
		Leakless(false)

	// Coba cari instalasi browser di sistem (Chrome/Edge/Chromium)
	// Jika ditemukan, gunakan binary tersebut alih-alih mendownload
	if path, exists := launcher.LookPath(); exists {
		l = l.Bin(path)
	}

	u := l.MustLaunch()

	// Connect ke browser
	s.browser = rod.New().ControlURL(u).MustConnect()
	return nil
}

// ScrapePage melakukan scraping sederhana pada halaman web
// url: Alamat web yang akan discrape
// selector: CSS selector untuk elemen yang ingin diambil teksnya (opsional)
func (s *BrowserService) ScrapePage(url string, selector string) ScrapeResult {
	if err := s.initBrowser(); err != nil {
		return ScrapeResult{Error: fmt.Sprintf("Failed to init browser: %v", err)}
	}

	// Buat page baru
	page := s.browser.MustPage(url)
	defer page.Close() // Pastikan page ditutup setelah selesai

	// Tunggu halaman selesai loading (Load Event Fired)
	// Kita juga bisa gunakan page.WaitStable() untuk menunggu animasi/ajax selesai
	if err := page.WaitLoad(); err != nil {
		return ScrapeResult{Error: fmt.Sprintf("Timeout waiting for load: %v", err)}
	}

	// Tunggu elemen spesifik jika diminta
	if selector != "" {
		// Timeout 5 detik untuk mencari elemen
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Race condition handling: Wait for element or timeout
		err := rod.Try(func() {
			page.Context(ctx).MustElement(selector)
		})
		if err != nil {
			return ScrapeResult{Error: fmt.Sprintf("Element selector '%s' not found", selector)}
		}
	}

	// Ambil data
	res := ScrapeResult{}

	// Ambil Title
	res.Title = page.MustInfo().Title

	// Ambil Content Text (jika ada selector)
	if selector != "" {
		if el, err := page.Element(selector); err == nil {
			res.Content = el.MustText()
		}
	} else {
		// Jika tidak ada selector, ambil seluruh body text
		res.Content = page.MustElement("body").MustText()
	}

	// Contoh: Ambil semua gambar di halaman
	// Evaluasi JS untuk mengambil src dari semua tag img
	// Ini contoh penggunaan MustEval
	imgList := page.MustEval(`() => Array.from(document.querySelectorAll('img')).map(i => i.src)`).Arr()
	for _, img := range imgList {
		if str, ok := img.Val().(string); ok && str != "" {
			res.Images = append(res.Images, str)
		}
	}

	return res
}

// Screenshot mengambil screenshot halaman
func (s *BrowserService) Screenshot(url string) (string, error) {
	if err := s.initBrowser(); err != nil {
		return "", err
	}

	page := s.browser.MustPage(url)
	defer page.Close()

	page.MustWaitLoad()

	// Ambil screenshot full page
	data, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format: proto.PageCaptureScreenshotFormatPng,
	})

	if err != nil {
		return "", err
	}

	// Return base64 string (biasanya butuh encoding, tapi di sini kita return raw bytes jika perlu,
	// namun untuk Wails biasanya base64 string lebih mudah dikirim ke frontend)
	// Untuk sederhananya, kita anggap return path atau base64 nanti.
	// Di sini saya hanya return pesan sukses dummy
	return fmt.Sprintf("Screenshot taken, size: %d bytes", len(data)), nil
}

// Cleanup menutup browser saat aplikasi berhenti
func (s *BrowserService) Cleanup() {
	if s.browser != nil {
		s.browser.MustClose()
	}
}
