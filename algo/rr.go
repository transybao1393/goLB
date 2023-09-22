package algo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"
)

var idx int = 0
var cfg Config

func init() {
	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(data, &cfg)
}

var retryBackends []RetryBackend

type RetryBackend struct {
	backend Backend
}

// Config is a configuration.
type Config struct {
	Proxy    Proxy     `json:"proxy"`
	Backends []Backend `json:"backends"`
}

// Proxy is a reverse proxy, and means load balancer.
type Proxy struct {
	Port string `json:"port"`
}

type Backend struct {
	URL    string `json:"url"`
	TYPE   string `json:"type"`
	RAM    string `json:"ram"`
	WEIGHT int    `json:"weight"`
	IsDead bool
	sync.RWMutex
}

// SetDead updates the value of IsDead in Backend.
func (backend *Backend) SetDead(b bool) {
	backend.Lock()
	backend.IsDead = b
	if b {
		//- Check if backend url is already in retryBackends
		//- if not, then append to retryBackends
		if len(retryBackends) > 0 {
			if backend.URL != retryBackends[len(retryBackends)-1].backend.URL {
				retryBackends = append(retryBackends, RetryBackend{backend: *backend})
			}
		}
		if len(retryBackends) == 0 {
			retryBackends = append(retryBackends, RetryBackend{backend: *backend})
		}
	}
	// fmt.Printf("\n>>>>>>> SET DEAD: %v value: %v - backend data: %+v\n\n", backend.URL, b, backend)
	backend.Unlock()
}

// GetIsDead returns the value of IsDead in Backend.
func (backend *Backend) GetIsDead() bool {
	backend.RLock()
	isDead := backend.IsDead
	fmt.Printf("\n>>>>>>> GetIsDead func is dead %t\n", isDead)
	backend.RUnlock()
	return isDead
}

type rr struct {
	items   []Backend
	count   int
	current int

	sync.Mutex
}

func NewRoundRobin() (lb *rr) {
	//- FIXME: Improve here
	//- TODO: This can be improved by receiving from many source
	var items [][]Backend
	items = append(items, cfg.Backends)
	lb = &rr{}
	if len(items) > 0 && len(items[0]) > 0 {
		lb.Update(items[0])
	}
	return
}

func (b *rr) Update(items interface{}) bool {
	v, ok := items.([]Backend)
	if !ok {
		return false
	}

	b.Lock()
	b.items = v
	b.count = len(v)
	b.current = idx
	b.Unlock()

	return true
}

func (b *rr) Select(_ ...string) (item string, currentBackend Backend) {
	b.Lock()

	b.current = b.current % b.count
	currentBackend = b.items[b.current]
	item = b.items[b.current].URL //- this value will be returned

	// fmt.Printf("\n>>>>>> retry backends %+v\n", retryBackends)

	idx++
	b.Unlock()

	return
}
