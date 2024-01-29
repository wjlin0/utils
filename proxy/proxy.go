package proxyutils

import (
	"context"
	"errors"
	"fmt"
	"github.com/remeh/sizedwaitgroup"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type proxyResult struct {
	AliveProxy string
	Error      error
}

var ProxyProbeConcurrency = 8

const (
	SOCKS5 = "socks5"
	HTTP   = "http"
	HTTPS  = "https"
)

// GetAnyAliveProxy returns a proxy from the list of proxies that is alive
func GetAnyAliveProxy(timeoutInSec int, proxies ...string) (string, error) {
	sg := sizedwaitgroup.New(ProxyProbeConcurrency)
	resChan := make(chan proxyResult, 4)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for _, v := range proxies {
			// skip iterating if alive proxy is found
			select {
			case <-ctx.Done():
				return
			default:
				proxy, err := GetProxyURL(v)
				if err != nil {
					resChan <- proxyResult{Error: err}
					continue
				}
				sg.Add()
				go func(proxyAddr url.URL) {
					defer sg.Done()
					select {
					case <-ctx.Done():
						return
					case resChan <- testProxyConn(proxyAddr, timeoutInSec):
						cancel()
					}
				}(proxy)
			}
		}
		sg.Wait()
		close(resChan)
	}()

	errstack := []string{}
	for {
		result, ok := <-resChan
		if !ok {
			break
		}
		if result.AliveProxy != "" {
			// found alive proxy return now
			return result.AliveProxy, nil
		} else if result.Error != nil {
			errstack = append(errstack, result.Error.Error())
		}
	}

	// all proxies are dead
	return "", fmt.Errorf("all proxies are dead got : %v", strings.Join(errstack, " : "))
}

// testProxyConn dial and test if proxy is open
func testProxyConn(proxyAddr url.URL, timeoutInSec int) proxyResult {
	p := proxyResult{}
	if Conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", proxyAddr.Hostname(), proxyAddr.Port()), time.Duration(timeoutInSec)*time.Second); err == nil {
		_ = Conn.Close()
		p.AliveProxy = proxyAddr.String()
	} else {
		p.Error = err
	}
	return p
}

// GetProxyURL returns a Proxy URL after validating if given proxy url is valid
func GetProxyURL(proxyAddr string) (url.URL, error) {
	if url, err := url.Parse(proxyAddr); err == nil && isSupportedProtocol(url.Scheme) {
		return *url, nil
	}
	return url.URL{}, errors.New("invalid proxy format (It should be http[s]/socks5://[username:password@]host:port)")
}

// isSupportedProtocol checks given protocols are supported
func isSupportedProtocol(value string) bool {
	return value == HTTP || value == HTTPS || value == SOCKS5
}

// GetProxyFunc returns a proxy func from the given proxy url
func GetProxyFunc(proxyURL string) (func(*http.Request) (*url.URL, error), error) {
	if proxyURL == "" {
		return nil, nil
	}
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}
	return http.ProxyURL(proxy), nil
}

// GetAnyAliveProxyFunc returns a proxy func from the given proxy url
func GetAnyAliveProxyFunc(timeoutInSec int, proxyURLs ...string) (func(*http.Request) (*url.URL, error), error) {
	var (
		err      error
		proxy    string
		proxyURL *url.URL
	)
	if len(proxyURLs) == 0 {
		return http.ProxyFromEnvironment, nil
	}

	proxy, err = GetAnyAliveProxy(timeoutInSec, proxyURLs...)
	if err != nil {
		return nil, err
	}
	proxyURL, err = url.Parse(proxy)
	if err != nil {
		return nil, err
	}

	return http.ProxyURL(proxyURL), nil
}
