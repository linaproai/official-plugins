// oauth_http.go holds shared HTTP helpers for OIDC outbound calls.

package oauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"lina-core/pkg/logger"
)

const (
	httpTimeout    = 15 * time.Second
	errorBodyLimit = 512
)

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return fmt.Sprintf("%s...(truncated)", value[:limit])
}

func closeResponseBody(ctx context.Context, resp *http.Response, endpointLabel string) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		logger.Warningf(ctx, "%s close response body failed err=%v", endpointLabel, err)
	}
}
