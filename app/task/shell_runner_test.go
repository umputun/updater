package task

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellRunner_Run(t *testing.T) {
	sr := ShellRunner{}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "echo 123", lw)
		t.Log(lw.String())
		require.NoError(t, err)
		assert.Equal(t, "123\n", lw.String())
	}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "no-such-command 123", lw)
		require.Error(t, err)
		t.Log(lw.String())
		assert.Contains(t, lw.String(), "not found")
	}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "@no-such-command 123", lw)
		t.Log(lw.String())
		require.NoError(t, err)
		assert.Contains(t, lw.String(), "not found")
	}

}

func TestShellRunner_RunMultiLine(t *testing.T) {
	sr := ShellRunner{}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "echo 123\necho 567\n", lw)
		require.NoError(t, err)
		assert.Equal(t, "123\n567\n", lw.String())
	}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "echo 123\nno-such-command 123", lw)
		require.Error(t, err)
		assert.Contains(t, lw.String(), "not found")
	}

	{
		lw := bytes.NewBuffer(nil)
		err := sr.Run(context.Background(), "echo 123\n@no-such-command 123", lw)
		require.NoError(t, err)
		assert.Contains(t, lw.String(), "not found")
	}

}
func TestShellRunner_RunBatch(t *testing.T) {
	sr := ShellRunner{BatchMode: true, TimeOut: time.Second}
	lw := bytes.NewBuffer(nil)
	err := sr.Run(context.Background(), "echo 123\necho 345", lw)
	require.NoError(t, err)
	assert.Equal(t, "123\n345\n", lw.String())
}

func TestShellRunner_RunBatchTimeOut(t *testing.T) {
	sr := ShellRunner{BatchMode: true, TimeOut: time.Millisecond * 100}
	lw := bytes.NewBuffer(nil)
	st := time.Now()
	err := sr.Run(context.Background(), "sleep 1 && sleep 1 && echo 123\necho 345", lw)
	require.Error(t, err)
	assert.True(t, time.Since(st) < time.Second*2)
}
