package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/updater/app/server/mocks"
)

func TestRest_Run(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srv := Rest{Listen: "localhost:54009", Version: "v1"}
	err := srv.Run(ctx)
	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestRest_taskCtrl(t *testing.T) {
	conf := &mocks.ConfigMock{GetTaskCommandFunc: func(name string) (string, bool) {
		return "echo " + name, true
	}}

	runner := &mocks.RunnerMock{RunFunc: func(context.Context, string, io.Writer) error {
		return nil
	}}

	srv := Rest{Listen: "localhost:54009", Version: "v1", Config: conf, SecretKey: "12345",
		Runner: runner, UpdateDelay: time.Millisecond * 200}

	ts := httptest.NewServer(srv.router())
	defer ts.Close()

	st := time.Now()
	resp, err := http.Get(ts.URL + "/update/task1/12345")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/update/task2/12345")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(ts.URL + "/update/task2/12345bad")
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.True(t, time.Since(st) >= time.Millisecond*200)
	assert.Equal(t, 2, len(conf.GetTaskCommandCalls()))
	assert.Equal(t, "task1", conf.GetTaskCommandCalls()[0].Name)
	assert.Equal(t, "task2", conf.GetTaskCommandCalls()[1].Name)

	assert.Equal(t, 2, len(runner.RunCalls()))
	assert.Equal(t, "echo task1", runner.RunCalls()[0].Command)
	assert.Equal(t, "echo task2", runner.RunCalls()[1].Command)
}

func TestRest_taskCtrlAsync(t *testing.T) {
	conf := &mocks.ConfigMock{GetTaskCommandFunc: func(name string) (string, bool) {
		return "echo " + name, true
	}}

	runner := &mocks.RunnerMock{RunFunc: func(context.Context, string, io.Writer) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}}

	srv := Rest{Listen: "localhost:54009", Version: "v1", Config: conf, SecretKey: "12345", Runner: runner}

	ts := httptest.NewServer(srv.router())
	defer ts.Close()

	st := time.Now()
	resp, err := http.Get(ts.URL + "/update/task1/12345?async=1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, time.Since(st) < 100*time.Millisecond, time.Since(st))
	time.Sleep(100 * time.Millisecond)
}

func TestRest_taskPostCtrl(t *testing.T) {
	conf := &mocks.ConfigMock{GetTaskCommandFunc: func(name string) (string, bool) {
		return "echo " + name, true
	}}

	runner := &mocks.RunnerMock{RunFunc: func(context.Context, string, io.Writer) error {
		return nil
	}}

	srv := Rest{Listen: "localhost:54009", Version: "v1", Config: conf, SecretKey: "12345",
		Runner: runner, UpdateDelay: time.Millisecond * 200}

	ts := httptest.NewServer(srv.router())
	defer ts.Close()

	st := time.Now()
	resp, err := http.Post(ts.URL+"/update", "application/json", strings.NewReader(`{"task":"task1","secret":"12345"}`))

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(ts.URL+"/update", "application/json", strings.NewReader(`{"task":"task2","secret":"12345"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Post(ts.URL+"/update", "application/json", strings.NewReader(`{"task":"task2","secret":"12345bad"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.True(t, time.Since(st) >= time.Millisecond*200)
	assert.Equal(t, 2, len(conf.GetTaskCommandCalls()))
	assert.Equal(t, "task1", conf.GetTaskCommandCalls()[0].Name)
	assert.Equal(t, "task2", conf.GetTaskCommandCalls()[1].Name)

	assert.Equal(t, 2, len(runner.RunCalls()))
	assert.Equal(t, "echo task1", runner.RunCalls()[0].Command)
	assert.Equal(t, "echo task2", runner.RunCalls()[1].Command)
}
