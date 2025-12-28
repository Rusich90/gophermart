package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type HTTPTestSetup struct {
	Server *httptest.Server
}

func StartHTTPServer(routes func(*gin.Engine)) *HTTPTestSetup {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	routes(r)

	server := httptest.NewServer(r)

	return &HTTPTestSetup{
		Server: server,
	}
}

func (h *HTTPTestSetup) PerformRequest(method, url string, body interface{}) (*http.Response, error) {
	var buf io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(b)
	}

	fullURL := h.Server.URL + url
	req, err := http.NewRequest(method, fullURL, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	return client.Do(req)
}

func (h *HTTPTestSetup) Close() {
	if h.Server != nil {
		h.Server.Close()
	}
}
