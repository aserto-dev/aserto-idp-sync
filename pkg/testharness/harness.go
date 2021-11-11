package testharness

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aserto-dev/go-utils/testutil"
	"github.com/stretchr/testify/require"

	"github.com/aserto-dev/idpsync/pkg/app"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
)

// TestHarness wraps a Idpsync so we can set it up easily
// and monitor its logs
type TestHarness struct {
	Idpsync    *app.App
	LogsReader *bufio.Reader

	cleanup      func()
	logFileRead  *os.File
	logFileWrite *os.File
	t            *testing.T
}

// Cleanup cleans up the application, releasing all resources
func (h *TestHarness) Cleanup() {
	assert := require.New(h.t)
	assert.NoError(h.Idpsync.Server.Stop())

	// Cleanup the app
	h.cleanup()
	assert.NoError(h.logFileRead.Close())
	assert.NoError(h.logFileWrite.Close())

	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8484")
	}, 10*time.Second, 10*time.Millisecond)
	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8383")
	}, 10*time.Second, 10*time.Millisecond)
	assert.Eventually(func() bool {
		return !testutil.PortOpen("127.0.0.1:8282")
	}, 10*time.Second, 10*time.Millisecond)
}

// Setup creates a new TestHarness
func Setup(t *testing.T, configOverrides func(*config.Config)) *TestHarness {
	assert := require.New(t)

	var err error
	h := &TestHarness{t: t}

	// Create a new pipe for writing our logs
	// we can use this stream to make assertions
	h.logFileRead, h.logFileWrite, err = os.Pipe()
	assert.NoError(err)

	h.LogsReader = bufio.NewReader(h.logFileRead)

	debugFileRead, debugFileWrite, err := os.Pipe()
	assert.NoError(err)

	logOut := ioutil.Discard
	if testutil.DebugFlagSet() {
		logOut = os.Stdout
	}
	debugReader := io.TeeReader(io.TeeReader(h.logFileRead, debugFileWrite), logOut)
	h.LogsReader = bufio.NewReader(debugFileRead)

	go func() {
		defer func() {
			assert.NoError(debugFileRead.Close())
			assert.NoError(debugFileWrite.Close())
		}()

		for {
			_, err := debugReader.Read(make([]byte, 1))

			if pErr, ok := err.(*os.PathError); ok {
				if pErr.Err == os.ErrClosed {
					break
				}
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				panic(err)
			}
		}
	}()

	h.Idpsync, h.cleanup, err = app.BuildTestIdpsync(
		h.logFileWrite, AssetDefaultConfig(), configOverrides)
	assert.NoError(err)

	err = h.Idpsync.Server.Start()
	assert.NoError(err)

	assert.Eventually(func() bool {
		return testutil.PortOpen("127.0.0.1:8383")
	}, 10*time.Second, 10*time.Millisecond)

	return h
}
