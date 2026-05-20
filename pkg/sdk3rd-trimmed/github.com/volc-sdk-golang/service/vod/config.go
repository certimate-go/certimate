package vod

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/volcengine/volc-sdk-golang/base"
	vevod "github.com/volcengine/volc-sdk-golang/service/vod"
)

type Vod struct {
	*base.Client
	DomainCache map[string]map[string]int
	Lock        sync.RWMutex
	disableLog  bool
}

type config struct {
	disableLog bool
}

type Option func(c *config)

func NewInstance(opts ...Option) *Vod {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	instance := &Vod{
		DomainCache: make(map[string]map[string]int),
		Client:      base.NewClient(ServiceInfoMap[base.RegionCnNorth1], ApiInfoList),
		disableLog:  cfg.disableLog,
	}
	return instance
}

func NewInstanceWithRegion(region string, opts ...Option) *Vod {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}
	var serviceInfo *base.ServiceInfo
	var ok bool
	if serviceInfo, ok = ServiceInfoMap[region]; !ok {
		serviceInfo = &base.ServiceInfo{
			Timeout: 60 * time.Second,
			Scheme:  "https",
			Host:    fmt.Sprintf("vod.%s.volcengineapi.com", region),
			Header: http.Header{
				"Accept": []string{"application/json"},
			},
			Credentials: base.Credentials{Region: region, Service: "vod"},
		}
	}

	instance := &Vod{
		DomainCache: make(map[string]map[string]int),
		Client:      base.NewClient(serviceInfo, ApiInfoList),
		disableLog:  cfg.disableLog,
	}
	return instance
}

var (
	ServiceInfoMap = vevod.ServiceInfoMap

	ApiInfoList = map[string]*base.ApiInfo{
		"ListDomain":         vevod.ApiInfoList["ListDomain"],
		"UpdateDomainConfig": vevod.ApiInfoList["UpdateDomainConfig"],
	}
)
