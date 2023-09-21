package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

var log = NewLogrusLogger()
var cfg Config

// - TODO: Handle the case that all backends are dead.
// - TODO: Implement retry when the backend is dead.
// - TODO: Implement healthcheck for backends.
// - TODO: Implement auto-reload for config.json.
// - TODO: gRPC & http/1.1 & QUIC support
//- TODO: Log all incomming requests.

// Config is a configuration.
type Config struct {
	Proxy    Proxy     `json:"proxy"`
	Backends []Backend `json:"backends"`
}

// Proxy is a reverse proxy, and means load balancer.
type Proxy struct {
	Port string `json:"port"`
}

// Backend is servers which load balancer is transferred.
type Backend struct {
	URL    string `json:"url"`
	TYPE   string `json:"type"`
	RAM    string `json:"ram"`
	IsDead bool
	mu     sync.RWMutex
}

// SetDead updates the value of IsDead in Backend.
func (backend *Backend) SetDead(b bool) {
	backend.mu.Lock()
	backend.IsDead = b
	backend.mu.Unlock()
}

// GetIsDead returns the value of IsDead in Backend.
func (backend *Backend) GetIsDead() bool {
	backend.mu.RLock()
	isAlive := backend.IsDead
	backend.mu.RUnlock()
	return isAlive
}

var mu sync.Mutex
var idx int = 0

// lbHandler is a handler for loadbalancing
func lbHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(Fields{
		"remote address": r.RemoteAddr,
		"datetime":       time.Now().Format("2006-01-02"),
		"method":         r.Method,
		"request uri":    r.RequestURI,
		"body":           r.Body,
		"header":         r.Header,
		"content length": r.ContentLength,
		"host":           r.Host,
		"proto":          r.Proto,
		"tls":            r.TLS,
	}).Info("Request information")

	maxLen := len(cfg.Backends)

	//- Round Robin
	//- FIXME: It is better to implement a better algorithm.
	mu.Lock()
	currentBackend := cfg.Backends[idx%maxLen]
	if currentBackend.GetIsDead() {
		idx++
	}
	targetURL, err := url.Parse(cfg.Backends[idx%maxLen].URL)
	if err != nil {
		log.Fatal(err, "Backend urls parse error")
	}
	idx++
	mu.Unlock()

	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		// NOTE: It is better to implement retry.
		log.Printf("%v is dead.", targetURL)
		currentBackend.SetDead(true)
		lbHandler(w, r)
	}
	reverseProxy.ServeHTTP(w, r)
}

// pingBackend checks if the backend is alive.
// - health check for every 1 minute
func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Unreachable to %v, error: %v", url.Host, err.Error())
		return false
	}
	defer conn.Close()
	return true
}

// healthCheck is a function for healthcheck
func healthCheck() {
	t := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-t.C:
			for _, backend := range cfg.Backends {
				pingURL, err := url.Parse(backend.URL)
				if err != nil {
					log.Fatal(err, "Ping error for healthcheck")
				}
				isAlive := isAlive(pingURL)
				backend.SetDead(!isAlive)
				msg := "OK"
				if !isAlive {
					msg = "DEAD"
				}
				// fmt.Printf("%v checked %v by healthcheck at %v\n", backend.URL, msg, currentTime)
				log.WithFields(Fields{
					"backend url": backend.URL,
					"status":      msg,
					"type":        backend.TYPE,
					"ram":         backend.RAM,
					"datetime":    time.Now().Format("2006-01-02"),
				}).Info("Healthchecking")

			}
		}
	}

}

// Serve serves a loadbalancer.
func Serve() {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal(err, "Cannot read config.json")
	}
	json.Unmarshal(data, &cfg)

	go healthCheck()

	fmt.Printf("Load Balancer is running on port %s...\n", cfg.Proxy.Port)
	s := http.Server{
		Addr:    ":" + cfg.Proxy.Port,
		Handler: http.HandlerFunc(lbHandler),
	}

	if err = s.ListenAndServe(); err != nil {
		log.Fatal(err, "Failed to listen and serve load balancer")
	}
}
