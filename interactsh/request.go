package interactshutils

import (
	"context"
	"github.com/projectdiscovery/interactsh/pkg/server"
)

type RequestData struct {
	MatchFunc   func(interactions *server.Interaction) bool
	ExtractFunc func(interactions *server.Interaction) []string
	match       bool
	extra       []string
	context     context.Context
	cancel      context.CancelFunc
}

func (r *RequestData) Result() (bool, []string) {

	<-r.context.Done()

	return r.match, r.extra
}

func DefaultHttpMatcher(interactions *server.Interaction) bool {
	return interactions.Protocol == "http"
}

func DefaultDnsMatcher(interactions *server.Interaction) bool {
	return interactions.Protocol == "dns"
}

func NewRequestData(matchFunc func(interactions *server.Interaction) bool, extractFunc func(interactions *server.Interaction) []string) *RequestData {
	return &RequestData{
		MatchFunc:   matchFunc,
		ExtractFunc: extractFunc,
	}
}

func NewDefaultHTTPMatcherRequestData() *RequestData {
	return NewRequestData(DefaultHttpMatcher, nil)
}

func NewDefaultDNSMatcherRequestData() *RequestData {
	return NewRequestData(DefaultDnsMatcher, nil)
}
