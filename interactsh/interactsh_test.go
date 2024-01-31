package interactshutils

import (
	"github.com/projectdiscovery/interactsh/pkg/server"
	"github.com/projectdiscovery/retryablehttp-go"
	"github.com/stretchr/testify/require"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestClient_Result(t *testing.T) {
	t.Run("TestClient_DefaultHTTPMatcherRequestData", func(t *testing.T) {
		opt := DefaultOptions(retryablehttp.DefaultHTTPClient)
		opt.PollDuration = 3 * time.Second
		client, err := New(opt)
		if err != nil {
			t.Error(err)
			return
		}
		var wg sync.WaitGroup
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func() {
				defer wg.Done()

				extractFunc := func(interactions *server.Interaction) []string {

					return nil
				}

				vul, _ := func() (bool, []string) {
					url, _ := client.NewURL()
					_, _ = http.DefaultClient.Get("http://" + url)

					req := NewDefaultHTTPMatcherRequestData()
					req.ExtractFunc = extractFunc
					client.ResultEventCallback(strings.Split(url, ".")[0], req)
					return req.Result()

				}()
				require.Equal(t, true, vul)
			}()
		}
		wg.Wait()

		client.Close()
		return
	})

	t.Run("TestClient_DefaultHTTPExtractorRequestData", func(t *testing.T) {
		opt := DefaultOptions(retryablehttp.DefaultHTTPClient)
		opt.PollDuration = 3 * time.Second
		client, err := New(opt)
		if err != nil {
			t.Error(err)
			return
		}
		var wg sync.WaitGroup
		wg.Add(10)
		for i := 0; i < 10; i++ {
			go func(i int) {
				defer wg.Done()

				regex, _ := regexp.Compile(`(?i)Foo: (.*)\r`)

				extractFunc := func(interactions *server.Interaction) []string {
					return regex.FindStringSubmatch(interactions.RawRequest)
				}

				_, ext := func() (bool, []string) {
					url, _ := client.NewURL()
					httpRequest, _ := retryablehttp.NewRequest(http.MethodGet, "http://"+url, nil)
					httpRequest.Header.Set("Foo", strconv.Itoa(i))
					_, _ = retryablehttp.DefaultHTTPClient.Do(httpRequest)
					req := &RequestData{
						MatchFunc:   DefaultHttpMatcher,
						ExtractFunc: extractFunc,
					}
					req.ExtractFunc = extractFunc
					client.ResultEventCallback(strings.Split(url, ".")[0], req)
					return req.Result()

				}()
				require.Equal(t, strconv.Itoa(i), ext[1])
			}(i)
		}
		wg.Wait()

		client.Close()
		return
	})
	return
}
