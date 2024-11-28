package icp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	proxyutils "github.com/wjlin0/utils/proxy"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const DefaultServer = "https://icp.wjlin0.com"

type Runner struct {
	client  *http.Client
	options *Options
}

func NewRunner(opts *Options) *Runner {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS10,
	}
	proxyFunc, _ := proxyutils.GetProxyFunc(opts.ProxyURL)
	if proxyFunc == nil {
		proxyFunc = http.ProxyFromEnvironment
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: -1,
		Proxy:               proxyFunc,
		TLSClientConfig:     tlsConfig,
	}
	if opts.Server == "" {
		opts.Server = DefaultServer
	}
	opts.Server = strings.TrimSuffix(opts.Server, "/")

	if opts.Retries < 0 {
		opts.Retries = 1
	}

	return &Runner{
		client: &http.Client{
			Timeout:       time.Second * time.Duration(30),
			CheckRedirect: nil,
			Transport:     transport,
		},
		options: opts,
	}
}

func (r *Runner) Search(keyword string) ([]*Entry, error) {
	retries := 0
	// 对 keyword 进行URL编码
	keyword = url.QueryEscape(keyword)
	req, err := http.NewRequest("GET", r.options.Server+"/query?type=cache&keyword="+keyword, nil)
	if err != nil {
		return nil, err
	}
xx:
	resp_, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp_.Body.Close()
	var resp *Response
	// json 序列化

	err = json.NewDecoder(resp_.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	// 判断是否成功
	if resp.Status != "successful" {
		if retries > r.options.Retries {
			retries++
			goto xx
		}
		return nil, fmt.Errorf("query failed: %s", resp.Message)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	return resp.Data, nil
}

// SearchByUnitName SearchBy 根据第一次搜索 unitName 获取所有的信息
func (r *Runner) SearchByUnitName(unitName string) ([]*Entry, error) {

	search, err := r.Search(unitName)
	if err != nil {
		return nil, err
	}
	var data []*Entry
	for _, entry := range search {
		search, err := r.Search(entry.UnitName)
		if err != nil {
			continue
		}
		data = append(data, search...)
	}
	return data, nil
}
