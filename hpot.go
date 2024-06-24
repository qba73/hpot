package hpot

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Pot represents a Honeypot.
type Pot struct {
	Verbose   bool
	AdminPort int
	Ports     []int

	mu      sync.Mutex
	records []net.Addr

	Log slog.Logger
}

// NewHoneyPotServer returns default, unstarted HoneyPot
// configured to listen for admin connections on port 8085.
func NewHoneyPotServer() *Pot {
	return &Pot{
		Verbose:   false,
		AdminPort: 8085,
		Log:       *slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

// NumConnections returns number of attempted connections to the Pot.
func (p *Pot) NumConnections() int {
	return len(p.Records())
}

// OpenPorts returns the ports Pot is listening on.
func (p *Pot) OpenPorts() []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Ports
}

// Records returns remote addresses representing incoming connections.
func (p *Pot) Records() []net.Addr {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.records
}

// ListenAndServe listens on the TCP network address addr on admin port
// and for upcoming network connections then calls serve with handler
// to handle requests on incoming connections on given ports.
func (p *Pot) ListenAndServe() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%d\n", p.NumConnections())
	})

	mux.HandleFunc("/ports", func(w http.ResponseWriter, r *http.Request) {
		openPorts := struct {
			OpenPorts []int `json:"openPorts"`
		}{
			OpenPorts: p.OpenPorts(),
		}
		err := json.NewEncoder(w).Encode(openPorts)
		if err != nil {
			return
		}
	})

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", p.AdminPort),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      mux,
	}

	go func() {
		p.Log.Info("starting admin server", "port", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	for _, port := range p.Ports {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		p.Log.Info("starting listener", "port", l.Addr().String())
		go p.serve(l)
	}
	// implement context
	select {}
}

// serve takes a listenr and starts listening for incomming connections.
func (p *Pot) serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
		p.Log.Info("incomming connection", "address", conn.RemoteAddr().String())
		p.mu.Lock()
		p.records = append(p.records, conn.RemoteAddr())
		p.mu.Unlock()
		conn.Close()
	}
}

func parsePorts(ports string) ([]int, error) {
	var prts []int
	px := strings.Split(ports, ",")
	for _, p := range px {
		p = strings.TrimSpace(p)
		pr, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		prts = append(prts, pr)
	}
	return prts, nil
}

var usage = `Usage: hpot [-v] [--admin=port] port1,port2,port3

Start the HopneyPot and listen on incoming connections on provided, comma-separated ports.

In verbose mode (-v), reports all incoming connections.`

func Main() int {
	verbose := flag.Bool("v", false, "verbose output")
	adminPort := flag.Int("admin", 8085, "admin port for reading metrics")
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println(usage)
		return 1
	}

	ports, err := parsePorts(flag.Args()[0])
	if err != nil {
		fmt.Println(usage)
		return 1
	}

	pot := NewHoneyPotServer()
	pot.Verbose = *verbose
	pot.AdminPort = *adminPort
	pot.Ports = ports

	if err := pot.ListenAndServe(); err != nil {
		return 1
	}
	return 0
}
