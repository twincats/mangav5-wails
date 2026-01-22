package downloader

import (
	"bytes"
	"strings"

	"github.com/go-resty/resty/v2"
)

func detectExt(resp *resty.Response) string {
	// 1️⃣ Content-Type header
	if ct := resp.Header().Get("Content-Type"); ct != "" {
		if ext := extFromContentType(ct); ext != "" {
			return ext
		}
	}

	body := resp.Body()
	if len(body) == 0 {
		return ".bin"
	}

	// 2️⃣ Magic byte
	if ext := extFromMagic(body); ext != "" {
		return ext
	}

	// 3️⃣ Final fallback
	return ".bin"
}

func extFromContentType(ct string) string {
	ct = strings.ToLower(ct)

	switch {
	case strings.Contains(ct, "jpeg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "webp"):
		return ".webp"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "avif"):
		return ".avif"
	default:
		return ""
	}
}

func extFromMagic(b []byte) string {
	switch {
	// JPEG
	case len(b) >= 3 &&
		b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return ".jpg"

	// PNG
	case len(b) >= 4 &&
		b[0] == 0x89 && b[1] == 0x50 &&
		b[2] == 0x4E && b[3] == 0x47:
		return ".png"

	// WEBP (RIFF....WEBP)
	case len(b) >= 12 &&
		string(b[0:4]) == "RIFF" &&
		string(b[8:12]) == "WEBP":
		return ".webp"

	// GIF
	case len(b) >= 4 &&
		string(b[0:4]) == "GIF8":
		return ".gif"

	// AVIF (ftypavif)
	case len(b) >= 16 &&
		bytes.Contains(b[:16], []byte("ftypavif")):
		return ".avif"

	default:
		return ""
	}
}
