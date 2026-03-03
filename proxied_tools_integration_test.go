//go:build integration

package mcpgrafana

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/grafana/grafana-openapi-client-go/client"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
)

// newTestContext creates a new context with the Grafana URL and service account token
// from the environment variables GRAFANA_URL and GRAFANA_SERVICE_ACCOUNT_TOKEN (or deprecated GRAFANA_API_KEY).
func newTestContext() context.Context {
	cfg := client.DefaultTransportConfig()
	cfg.Host = "localhost:3000"
	cfg.Schemes = []string{"http"}
	// Extract transport config from env vars, and set it on the context.
	if u, ok := os.LookupEnv("GRAFANA_URL"); ok {
		url, err := url.Parse(u)
		if err != nil {
			panic(fmt.Errorf("invalid %s: %w", "GRAFANA_URL", err))
		}
		cfg.Host = url.Host
		// The Grafana client will always prefer HTTPS even if the URL is HTTP,
		// so we need to limit the schemes to HTTP if the URL is HTTP.
		if url.Scheme == "http" {
			cfg.Schemes = []string{"http"}
		}
	}

	// Check for the new service account token environment variable first
	if apiKey := os.Getenv("GRAFANA_SERVICE_ACCOUNT_TOKEN"); apiKey != "" {
		cfg.APIKey = apiKey
	} else if apiKey := os.Getenv("GRAFANA_API_KEY"); apiKey != "" {
		// Fall back to the deprecated API key environment variable
		cfg.APIKey = apiKey
	} else {
		cfg.BasicAuth = url.UserPassword("admin", "admin")
	}

	client := client.NewHTTPClientWithConfig(strfmt.Default, cfg)

	grafanaCfg := GrafanaConfig{
		Debug:     true,
		URL:       "http://localhost:3000",
		APIKey:    cfg.APIKey,
		BasicAuth: cfg.BasicAuth,
	}

	ctx := WithGrafanaConfig(context.Background(), grafanaCfg)
	return WithGrafanaClient(ctx, client)
}

func Test_ProxyToolsConcurrency(t *testing.T) {

	//todo : change the test to validate incoming notification count to be equal to one 
	t.Run("concurrent list tool calls", func(t *testing.T) {
		//should fire list change notification only once
		ctx := newTestContext()
		sm := NewSessionManager()

		//list tools or tool call
		mcpServer := server.NewMCPServer("test", "1.0.0")
		mockSession := &mockClientSession{id: "session-mock"}
		mcpServer.RegisterSession(ctx, mockSession)
		tm := NewToolManager(sm, mcpServer, WithProxiedTools(true))
		for range 10 {
			tm.InitializeAndRegisterProxiedTools(ctx, mockSession)
		}

		time.Sleep(2 * time.Second)
		msgCount := len(mockSession.NotificationChannel())

		assert.Equal(t, 1, msgCount, "should send notification only once")

	})
}
