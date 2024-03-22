package hpot_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestPotStervesHTTPStatusPage(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, t *http.Request) {
		fmt.Fprint(w, "Honeypot stats")
	}))
	defer ts.Close()

	client := ts.Client()

	// user behaviour
	res, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatal(res.StatusCode)
	}

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	want := "Honeypot stats"
	if want != string(got) {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestPotStervesStatisticsWithoutConnections(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, t *http.Request) {
		fmt.Fprint(w, potStatsWithoutConnections)
	}))
	defer ts.Close()

	client := ts.Client()

	// user behaviour
	res, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatal(res.StatusCode)
	}

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var ps hpot.PotStats
	if err := json.Unmarshal(got, &ps); err != nil {
		t.Fatal(err)
	}

	want := hpot.PotStats{
		Connections: 0,
	}

	if !cmp.Equal(want, ps) {
		t.Errorf(cmp.Diff(want, ps))
	}
}

func TestPotStervesStatisticsWithConnections(t *testing.T) {
	t.Parallel()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, t *http.Request) {
		fmt.Fprint(w, potStatsWithConnections)
	}))
	defer ts.Close()

	client := ts.Client()

	// user behaviour
	res, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatal(res.StatusCode)
	}

	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var ps hpot.PotStats
	if err := json.Unmarshal(got, &ps); err != nil {
		t.Fatal(err)
	}

	want := hpot.PotStats{
		Connections: 2,
	}

	if !cmp.Equal(want, ps) {
		t.Errorf(cmp.Diff(want, ps))
	}
}

var (
	potStatsWithoutConnections = `{"connections": 0}`
	potStatsWithConnections    = `{"connections": 2}`
)
