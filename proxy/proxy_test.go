//go:build proxy

package proxyutils

// package tests will be executed only with (running proxy is necessary):
// go test -tags proxy

import (
	"crypto/tls"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func TestGetAnyAliveProxyFunc(t *testing.T) {
	// Test proxy
	proxys := []string{
		"socks5://127.0.0.1:7890",
	}
	proxyFunc, err := GetAnyAliveProxyFunc(5, proxys...)
	if err != nil {
		t.Error(err)
		return
	}
	proxyClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: proxyFunc,
		},
	}
	noProxyclient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	proxyResponse, err := proxyClient.Get("https://ifconfig.me/ip")
	if err != nil {
		t.Error(err)
		return
	}
	defer proxyResponse.Body.Close()
	noProxyResponse, err := noProxyclient.Get("https://ifconfig.me/ip")
	if err != nil {
		t.Error(err)
		return
	}
	defer noProxyResponse.Body.Close()
	noProxyIp, _ := io.ReadAll(noProxyResponse.Body)
	proxyIp, _ := io.ReadAll(proxyResponse.Body)
	require.NotEmpty(t, string(noProxyIp))
	require.NotEmpty(t, string(proxyIp))
	require.NotEqual(t, string(noProxyIp), string(proxyIp))
}
