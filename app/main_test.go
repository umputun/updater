package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_main(t *testing.T) {
	var buf [4]byte
	_, err := rand.Read(buf[:])
	require.NoError(t, err)
	port := 40000 + int(binary.BigEndian.Uint32(buf[:])%10000)
	os.Args = []string{"app", "--key=12345", "--listen=127.0.0.1:" + strconv.Itoa(port), "--file=../updater.yml", "--dbg"}

	done := make(chan struct{})
	go func() {
		<-done
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/update/test-echo/12345", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}

	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/update/test-echo/bad12345", port))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, 403, resp.StatusCode)
	}
}

func waitForHTTPServerStart(port int) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{Timeout: time.Second}
	for range 100 {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}
