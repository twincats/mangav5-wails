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
	browser  *rod.Browser
	launcher *launcher.Launcher
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
		Leakless(false).
		Set("disable-web-security", "true").         // Disable CORS
		Set("disable-site-isolation-trials", "true") // Disable site isolation

	// Coba cari instalasi browser di sistem (Chrome/Edge/Chromium)
	// Jika ditemukan, gunakan binary tersebut alih-alih mendownload
	if path, exists := launcher.LookPath(); exists {
		l = l.Bin(path)
	}

	// Simpan instance launcher untuk keperluan cleanup
	s.launcher = l

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
	if info, err := page.Info(); err == nil {
		res.Title = info.Title
	}

	// Ambil Content berdasarkan selector
	if selector != "" {
		if el, err := page.Element(selector); err == nil {
			if text, err := el.Text(); err == nil {
				res.Content = text
			}
		}
	} else {
		// Default ambil body text jika tidak ada selector
		if el, err := page.Element("body"); err == nil {
			if text, err := el.Text(); err == nil {
				res.Content = text
			}
		}
	}

	// Contoh: Ambil semua gambar di halaman
	// Evaluasi JS untuk mengambil src dari semua tag img
	imgList, err := page.Eval(`() => Array.from(document.querySelectorAll('img')).map(i => i.src)`)
	if err == nil {
		for _, img := range imgList.Value.Arr() {
			if str, ok := img.Val().(string); ok && str != "" {
				res.Images = append(res.Images, str)
			}
		}
	}

	return res
}

// Close menutup browser instance
func (s *BrowserService) Close() error {
	if s.browser != nil {
		return s.browser.Close()
	}
	return nil
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

// ScrapFull mengambil seluruh HTML halaman setelah render selesai
func (s *BrowserService) ScrapFull(url string) (string, error) {
	if err := s.initBrowser(); err != nil {
		return "", fmt.Errorf("failed to init browser: %w", err)
	}
	page := s.browser.MustPage(url)
	defer page.Close()

	// Tunggu hingga halaman stabil (network idle & tidak ada perubahan DOM)
	// Ini penting untuk website SPA atau yang menggunakan banyak JS
	page.MustWaitStable()

	return page.MustHTML(), nil
}

// Cleanup menutup browser dan membersihkan resource
func (s *BrowserService) Cleanup() {
	if s.browser != nil {
		_ = s.browser.Close()
		s.browser = nil
	}

	// Matikan proses browser via launcher
	if s.launcher != nil {
		s.launcher.Kill()
		s.launcher = nil
	}
}
