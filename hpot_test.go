package hpot_test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/hpot"
)

func TestHoneypotAcceptsConection(t *testing.T) {
	t.Parallel()

	port1 := randomFreePort()
	port2 := randomFreePort()

	pot := hpot.NewHoneyPotServer()
	pot.AdminPort = randomFreePort()
	pot.Ports = []int{port1, port2}

	go func() {
		if err := pot.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	client1 := mustConnect(t, port1)
	client2 := mustConnect(t, port2)
	client3 := mustConnect(t, port1)
	client4 := mustConnect(t, port2)

	want := []net.Addr{client1, client2, client3, client4}

	got := pot.Records()

	// wait for success
	for len(got) < 4 {
		time.Sleep(10 * time.Millisecond)
		got = pot.Records()
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestPotReturnsConnectionCount(t *testing.T) {
	t.Parallel()

	port := randomFreePort()
	pot := hpot.NewHoneyPotServer()
	pot.AdminPort = randomFreePort()
	pot.Ports = []int{port}

	go func() {
		if err := pot.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	got := getStats(t, pot.AdminPort)
	want := "0\n"

	if want != got {
		t.Error(cmp.Diff(want, got))
	}

	mustConnect(t, port)
	mustConnect(t, port)

	got = getStats(t, pot.AdminPort)
	want = "2\n"

	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func TestPotReturnOpenPorts(t *testing.T) {
	t.Parallel()

	port := randomFreePort()
	pot := hpot.NewHoneyPotServer()
	pot.AdminPort = randomFreePort()
	pot.Ports = []int{port}

	go func() {
		if err := pot.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	got := getOpenPorts(t, pot.AdminPort)
	want := fmt.Sprintf("[%d]\n", port)

	if want != got {
		t.Error(cmp.Diff(want, got))
	}
}

func getOpenPorts(t *testing.T, port int) string {
	t.Helper()

	url := fmt.Sprintf("http://127.0.0.1:%d/ports", port)
	res, err := http.Get(url)
	for err != nil {
		time.Sleep(20 * time.Millisecond)
		res, err = http.Get(url)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatal(res.StatusCode)
	}

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)

	}
	return string(got)
}

func getStats(t *testing.T, port int) string {
	t.Helper()

	url := fmt.Sprintf("http://127.0.0.1:%d/metrics", port)
	res, err := http.Get(url)
	for err != nil {
		time.Sleep(20 * time.Millisecond)
		res, err = http.Get(url)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatal(res.StatusCode)
	}

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)

	}
	return string(got)
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

	for {
		conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		defer conn.Close()
		return conn.LocalAddr()
	}
}
