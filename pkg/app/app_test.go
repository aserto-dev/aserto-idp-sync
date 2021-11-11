package app_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/aserto-dev/idpsync/api/idpsync/v1"
	"github.com/aserto-dev/idpsync/pkg/cc/config"
	"github.com/aserto-dev/idpsync/pkg/testharness"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestSyncEndpoint(t *testing.T) {
	// Arrange
	h := testharness.Setup(t, func(cfg *config.Config) {})
	defer h.Cleanup()
	assert := require.New(t)

	// Act
	client := h.CreateClient()
	url := "https://127.0.0.1:8383/api/v1/sync/user"

	syncReq := idpsync.SyncUserRequest{
		EmailAddress: "",
	}

	buf, err := protojson.Marshal(&syncReq)
	assert.NoError(err)

	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	assert.NoError(err)

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)

	assert.Equal(400, resp.StatusCode)
	_ = body
}
