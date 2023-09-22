package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"load-balancer/algo"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// - TODO: Handle the case that all backends are dead.
// - TODO: Implement retry when the backend is dead. => notify administrators
// - TODO: Implement auto-reload for config.json.
// - TODO: gRPC & http/1.1 & QUIC support
// - TODO: Basic security (Implement SSL/TLS encryption, Apply firewall rules and access control lists, Use authentication and authorization)
// - TODO: Implement more Load Balancing Algorithms (Least Connections, Least Time, Hash, IP Hash, Random with Two Choices,...)

func lbHandlerRRv1(w http.ResponseWriter, r *http.Request) {
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

	//- RR version 1
	currentBackend, targetURL, _ := algo.SelectRRv1()

	//- check isAlive before serving, 1 second
	if isAliveWithTimeLimit(targetURL, 1*time.Second) {
		//- if alive
		currentBackend.SetDead(false)
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			// - TODO: It is better to implement retry
			log.WithFields(Fields{
				"status":  "WARNING",
				"server":  targetURL,
				"message": fmt.Sprintf("Cannot establish request to %s\n", targetURL),
			}).Info("REQUEST ESTABLISH")
		}
		reverseProxy.ServeHTTP(w, r)
	} else {
		//- if not, add to retry backends
		currentBackend.SetDead(true)
		log.WithFields(Fields{
			"status":  "DEAD",
			"server":  targetURL,
			"message": fmt.Sprintf("Server with URL %v is DEAD.", targetURL),
		}).Info("SERVER DEAD")
		lbHandlerRRv1(w, r) //- retry
	}
}

func lbHandlerRRv2(w http.ResponseWriter, r *http.Request) {
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

	//- RR version 2
	// mu.Lock()
	lb := algo.NewRoundRobin()

	nextBackendURL, currentBackend := lb.Select()
	targetURL, err := url.Parse(nextBackendURL)
	fmt.Printf("currentBackendURL %s\n", nextBackendURL)
	if err != nil {
		log.Fatal(err, "Backend urls parse error")
	}
	// mu.Unlock()

	//- check isAlive before serving, 1 second
	if isAliveWithTimeLimit(targetURL, 1*time.Second) {
		//- if alive
		currentBackend.SetDead(false)
		reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
		reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			// - TODO: It is better to implement retry
			log.WithFields(Fields{
				"status":  "WARNING",
				"server":  targetURL,
				"message": fmt.Sprintf("Cannot establish request to %s\n", targetURL),
			}).Info("REQUEST ESTABLISH")
		}
		reverseProxy.ServeHTTP(w, r)
	} else {
		//- if not, add to retry backends
		currentBackend.SetDead(true)
		log.WithFields(Fields{
			"status":  "DEAD",
			"server":  targetURL,
			"message": fmt.Sprintf("Server with URL %v is DEAD.", targetURL),
		}).Info("SERVER DEAD")
		lbHandlerRRv2(w, r) //- retry
	}
}

// pingBackend checks if the backend is alive.
// - is alive with default 1 minute
func isAlive(url *url.URL) bool {
	conn, err := net.DialTimeout("tcp", url.Host, time.Minute*1)
	if err != nil {
		log.Printf("Unreachable to %v, error: %v", url.Host, err.Error())
		return false
	}
	defer conn.Close()
	return true
}

// - is alive with time limit, maybe a few seconds
func isAliveWithTimeLimit(url *url.URL, t time.Duration) bool {
	conn, err := net.DialTimeout("tcp", url.Host, t)
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
		Handler: http.HandlerFunc(lbHandlerRRv1),
	}

	if err = s.ListenAndServe(); err != nil {
		log.Fatal(err, "Failed to listen and serve load balancer")
	}
}
