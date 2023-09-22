package algo

import (
	"fmt"
	"net/url"
	"sync"
)

var mu sync.Mutex
var idxv1 int = 0

// Proxy is a reverse proxy, and means load balancer.
type Proxyv1 struct {
	Port string `json:"port"`
}

// - Select next backend URL, return next URL
func SelectRRv1() (*Backend, *url.URL, error) {
	maxLen := len(cfg.Backends)

	mu.Lock()
	currentBackend := cfg.Backends[idxv1%maxLen] //- Round Robin calculate

	targetURL, err := url.Parse(cfg.Backends[idxv1%maxLen].URL) //- calculate again incase backend url is dead
	fmt.Printf("URL:  %s\n", cfg.Backends[idxv1%maxLen].URL)
	if err != nil {
		return nil, &url.URL{}, err
	}

	// fmt.Printf("\n>>>>>> retry backends %+v\n", retryBackends)

	idxv1++
	mu.Unlock()

	return &currentBackend, targetURL, nil
}
