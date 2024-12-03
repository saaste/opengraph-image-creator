package image

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

func TakeScreenshot(url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Creating timeout for 15 seconds
	ctx, cancel = context.WithTimeout(ctx, time.Second*15)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		emulation.SetUserAgentOverride("OpenGraphImageCreator"),
		chromedp.Navigate(url),
		chromedp.Screenshot(".opengraph", &buf, chromedp.NodeVisible),
	)

	if err != nil {
		return buf, err
	}

	return buf, nil
}
