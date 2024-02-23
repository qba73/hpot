package hpot_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/hpot"
)

func TestHoneypotAcceptsConection(t *testing.T) {
	t.Parallel()

	port1 := randomFreePort()
	port2 := randomFreePort()

	pot, err := hpot.StartHoneypotOnPorts(false, port1, port2)
	if err != nil {
		t.Fatal(err)
	}

	client1 := mustConnect(t, port1)
	client2 := mustConnect(t, port2)

	client3 := mustConnect(t, port1)
	client4 := mustConnect(t, port2)

	want := []net.Addr{client1, client2, client3, client4}

	// Fix race conditions here - check ports!
	// temp hack!!!
	time.Sleep(2 * time.Second)

	got := pot.Records()

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func randomFreePort() int {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	tcpAddr, err := net.ResolveTCPAddr("tcp", l.Addr().String())
	if err != nil {
		panic(err)
	}
	return tcpAddr.Port
}

func mustConnect(t *testing.T, port int) net.Addr {
	t.Helper()
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	return conn.LocalAddr()
}
