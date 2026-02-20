package utils

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/emon51/rental-scraper/config"
)

// CreateBrowserContext creates and configures the browser context
func CreateBrowserContext(cfg *config.Config) context.Context {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, _ = context.WithTimeout(ctx, time.Duration(cfg.PageTimeout)*time.Second)

	return ctx
}