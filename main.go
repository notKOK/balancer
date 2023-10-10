package main

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

type Config struct {
	ServiceUrl  []string `yaml:"service_url"`
	BalancePort int      `yaml:"balance_port"`
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IsAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

type ServerPool struct {
	backends []*Backend
	current  uint64
}

var configPath = "./.env/config.yml"

func GetConfig() *Config {
	instance := &Config{}
	err := cleanenv.ReadConfig(configPath, instance)
	if err != nil {
		log.Fatal(err)
	}
	return instance
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}

func lb(w http.ResponseWriter, r *http.Request) {
	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)
	return true
}

func healthCheck() {
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			serverPool.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

var serverPool ServerPool

func main() {
	config := GetConfig()

	if len(config.ServiceUrl) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	for _, tok := range config.ServiceUrl {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}
		serverPool.backends = append(serverPool.backends, &Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: httputil.NewSingleHostReverseProxy(serverUrl),
		})
		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", config.BalancePort),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	log.Printf("Load Balancer started at :%d\n", config.BalancePort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
