package api

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPServerBoundsIncompleteLoginBodyRead(t *testing.T) {
	router, _ := newLimitedLoginRouter(t)
	api := &Api{Router: router}
	server := api.httpServer("127.0.0.1:0")
	if server.Addr != "127.0.0.1:0" || server.Handler != router || server.ReadTimeout != 30*time.Second || server.ReadHeaderTimeout != 5*time.Second {
		t.Fatal("HTTP server configuration did not retain address, handler, and finite read bounds")
	}
	// Shorten only the fixture deadline; exercise the real net/http timeout while
	// the real Gin middleware waits for the unfinished login JSON.
	server.ReadTimeout = 100 * time.Millisecond
	server.ReadHeaderTimeout = 100 * time.Millisecond
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	stopped := make(chan error, 1)
	go func() { stopped <- server.Serve(listener) }()
	t.Cleanup(func() {
		_ = server.Close()
		if err := <-stopped; err != http.ErrServerClosed {
			t.Errorf("HTTP server stopped unexpectedly: %v", err)
		}
	})
	connection, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	_, err = fmt.Fprint(connection, "POST /login HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\nConnection: close\r\n\r\n{\"type\":\"login\",")
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.ReadResponse(bufio.NewReader(connection), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "i/o timeout") {
		t.Fatalf("incomplete JSON did not reach bounded existing bind-error path: HTTP %d %s", response.StatusCode, body)
	}
}
