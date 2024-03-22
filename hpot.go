package hpot

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

// Pot represents a Honeypot.
type Pot struct {
	Verbose   bool
	AdminPort int
	Ports     []int

	mu      sync.Mutex
	records []net.Addr
}

// NewHoneyPotServer returns default, unstarted HoneyPot
// configured to listen for admin connections on port 8085.
func NewHoneyPotServer() *Pot {
	return &Pot{
		Verbose:   false,
		AdminPort: 8085,
	}
}

// NumConnections returns number of attempted connections to the Pot.
func (p *Pot) NumConnections() int {
	return len(p.Records())
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
	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%d", p.NumConnections())
	})

	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", p.AdminPort),
		Handler: mux,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	for _, port := range p.Ports {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		if p.Verbose {
			fmt.Println("Starting listener on:", l.Addr().String())
		}
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
		if p.Verbose {
			fmt.Println("Incomming connection from: ", conn.RemoteAddr())
		}
		p.mu.Lock()
		p.records = append(p.records, conn.RemoteAddr())
		p.mu.Unlock()

		conn.Close()
	}
}
