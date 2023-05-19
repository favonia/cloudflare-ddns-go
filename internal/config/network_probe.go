package config

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/favonia/cloudflare-ddns/internal/pp"
)

// ProbeURL quickly checks whether one can send a HEAD request to the url.
func ProbeURL(ctx context.Context, url string) bool {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	return err == nil
}

// ShouldWeUse1001 quickly checks 1.1.1.1 and 1.0.0.1 and return whether 1.0.0.1 should be used.
func ShouldWeUse1001(ctx context.Context, ppfmt pp.PP) bool {
	if !ProbeURL(ctx, "https://1.1.1.1") && ProbeURL(ctx, "https://1.0.0.1") {
		ppfmt.Warningf(pp.EmojiError, "1.1.1.1 appears to be blocked or intercepted by your ISP or your router")
		ppfmt.Warningf(pp.EmojiGood, "1.0.0.1 seems to work and will be used instead of 1.1.1.1 for IPv4 address detection")
		return true
	}

	return false
}
