package downloader

import (
	"time"

	"github.com/go-resty/resty/v2"
)

func NewRestyClient(timeout time.Duration) *resty.Client {
	return resty.New().
		SetTimeout(timeout).
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(5)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
}
