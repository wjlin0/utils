package interactshutils

import (
	"context"
	"github.com/Mzack9999/gcache"
	"github.com/projectdiscovery/interactsh/pkg/client"
	"github.com/projectdiscovery/interactsh/pkg/server"
	errorutil "github.com/projectdiscovery/utils/errors"
	"sync"
	"time"
)

type Client struct {
	interactsh   *client.Client
	interactions gcache.Cache[string, []*server.Interaction]
	requestDatas gcache.Cache[string, *RequestData]
	options      *Options
	sync.Once
	sync.RWMutex
	pollDuration     time.Duration
	cooldownDuration time.Duration
	eviction         time.Duration
}

// New returns a new interactsh server client
func New(options *Options) (*Client, error) {
	interactionsCache := gcache.New[string, []*server.Interaction](defaultMaxInteractionsCount).LRU().Build()
	requestDataCache := gcache.New[string, *RequestData](defaultMaxInteractionsCount).LRU().Build()
	interactClient := &Client{
		eviction:     options.Eviction,
		interactions: interactionsCache,
		options:      options,
		requestDatas: requestDataCache,
		pollDuration: options.PollDuration,
	}

	return interactClient, nil
}

func (c *Client) NewURL() (string, error) {
	var (
		err error
	)
	c.Do(func() {
		err = c.poll()
	})
	if err != nil {
		return "", ErrInteractshClientNotInitialized
	}

	if c.interactsh == nil {
		return "", ErrInteractshClientNotInitialized
	}

	url := c.interactsh.URL()

	return url, nil

}
func (c *Client) poll() error {
	interactsh, err := client.New(&client.Options{
		ServerURL:           c.options.ServerURL,
		Token:               c.options.Token,
		DisableHTTPFallback: c.options.DisableHttpFallback,
		HTTPClient:          c.options.HTTPClient,
		KeepAliveInterval:   time.Minute,
	})
	if err != nil {
		return err
	}
	c.interactsh = interactsh

	err = interactsh.StartPolling(c.pollDuration, func(interaction *server.Interaction) {
		request, err := c.requestDatas.Get(interaction.UniqueID)
		if errorutil.IsAny(err, gcache.KeyNotFoundError) || request == nil {
			items, err := c.interactions.Get(interaction.UniqueID)
			if errorutil.IsAny(err, gcache.KeyNotFoundError) || items == nil {
				_ = c.interactions.SetWithExpire(interaction.UniqueID, []*server.Interaction{interaction}, defaultInteractionDuration)
			} else {
				items = append(items, interaction)
				_ = c.interactions.SetWithExpire(interaction.UniqueID, items, defaultInteractionDuration)
			}
			return
		}

		c.processInteractionForRequest(interaction, request)
	})

	if err != nil {
		return errorutil.NewWithErr(err).Msgf("could not perform interactsh polling")
	}
	return err
}
func (c *Client) ResultEventCallback(id string, data *RequestData) {

	interactions, err := c.interactions.Get(id)
	if interactions != nil && err == nil {
		for _, interaction := range interactions {
			if c.processInteractionForRequest(interaction, data) {
				c.interactions.Remove(id)
				break
			}
		}
	} else {
		data.context, data.cancel = context.WithTimeout(context.Background(), c.pollDuration)
		_ = c.requestDatas.SetWithExpire(id, data, c.eviction)
	}

	return
}

func (c *Client) Close() {
	c.interactions.Purge()
	if c.interactsh != nil {
		_ = c.interactsh.StopPolling()
		_ = c.interactsh.Close()
	}
	return
}
func (c *Client) processInteractionForRequest(interaction *server.Interaction, data *RequestData) bool {
	var (
		extra []string
		match bool
	)
	c.Lock()
	if data.MatchFunc != nil {
		match = data.MatchFunc(interaction)
	}
	c.Unlock()
	if !match {
		return false
	}

	c.Lock()
	if data.ExtractFunc != nil {
		extra = data.ExtractFunc(interaction)
	}

	c.Unlock()
	data.match = match
	data.extra = extra
	data.cancel()

	return true
}
