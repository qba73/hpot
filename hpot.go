package hpot

import (
	"fmt"
	"net"
	"os"
	"sync"
)

// Pot represents a Honeypot.
type Pot struct {
	verbose bool

	mu      sync.Mutex
	records []net.Addr
}

// Records returns remote addresses representing incoming connections.
func (p *Pot) Records() []net.Addr {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.records
}

// serve takes a listenr and starts listening for incomming connections.
func (p *Pot) serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stdout, err)
		}
		if p.verbose {
			fmt.Println("Incomming connection from: ", conn.RemoteAddr())
		}
		p.mu.Lock()
		p.records = append(p.records, conn.RemoteAddr())
		p.mu.Unlock()

		conn.Close()
	}
}

// StartHoneypotOnPorts starts the honeypot on given TCP ports.
// If verbose is set to `true` honeypot will log listening ports
// and IP addresses of incomming connections.
func StartHoneypotOnPorts(verbose bool, ports ...int) (*Pot, error) {
	p := &Pot{
		verbose: verbose,
	}
	for _, port := range ports {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return nil, err
		}
		if p.verbose {
			fmt.Println("Starting listener on:", l.Addr().String())
		}
		go p.serve(l)
	}
	return p, nil
}
