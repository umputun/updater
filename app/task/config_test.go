package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	c, err := LoadConfig("testdata/test.yml")
	require.NoError(t, err)
	assert.Equal(t, 2, len(c.Tasks))
	assert.Equal(t, "test1", c.Tasks[0].Name)
	assert.Equal(t, "test2", c.Tasks[1].Name)
	assert.Equal(t, "do blah1", c.Tasks[0].Command)
	assert.Equal(t, "do blah2", c.Tasks[1].Command)

	_, err = LoadConfig("no-such-file.yml")
	assert.Error(t, err)
}

func TestConfig_GetTaskCommand(t *testing.T) {
	c, err := LoadConfig("testdata/test.yml")
	require.NoError(t, err)

	r, ok := c.GetTaskCommand("test1")
	require.True(t, ok)
	assert.Equal(t, "do blah1", r)

	r, ok = c.GetTaskCommand("test2")
	require.True(t, ok)
	assert.Equal(t, "do blah2", r)

	_, ok = c.GetTaskCommand("bad-task")
	require.False(t, ok)
}
