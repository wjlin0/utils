package interactshutils

import (
	"github.com/projectdiscovery/interactsh/pkg/client"
	"github.com/projectdiscovery/retryablehttp-go"
	"time"
)

type Options struct {
	ServerURL           string
	Token               string
	HTTPClient          *retryablehttp.Client
	CacheSize           int
	Eviction            time.Duration
	PollDuration        time.Duration
	DisableHttpFallback bool
}

func DefaultOptions(httpClient *retryablehttp.Client) *Options {
	return &Options{
		HTTPClient:          httpClient,
		DisableHttpFallback: false,
		ServerURL:           client.DefaultOptions.ServerURL,
		CacheSize:           5000,
		Eviction:            60 * time.Second,
		PollDuration:        5 * time.Second,
	}
}
