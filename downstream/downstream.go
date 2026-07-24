package downstream

import (
	"context"
	"io"
	"log"
	"log/slog"
	"net/http"

	cModel "github.com/ygo-skc/skc-go/common/v3/model"

	"github.com/ygo-skc/skc-go/common/v3/client"
	cUtil "github.com/ygo-skc/skc-go/common/v3/util"
)

var (
	YGO client.YGOClientImpV1
)

func ConnectToYGOService() {
	if c, err := client.NewYGOServiceClients("ygo-service.skc.cards", cUtil.EnvMap["YGO_SERVICE_HOST"]); err != nil {
		log.Fatalf("Failed to connect to ygo-service: %v", err)
	} else {
		YGO = *c
	}
}

func parseResponseBody(ctx context.Context, resp *http.Response) ([]byte, *cModel.APIError) {
	logger := cUtil.RetrieveLogger(ctx)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading downstream response body", slog.Any("err", err), slog.Any("url", resp.Request.URL))
		return nil, &cModel.APIError{Message: "Error reading response from downstream service", StatusCode: http.StatusInternalServerError}
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Downstream service returned non-200 response", slog.Int("status", resp.StatusCode), slog.Any("url", resp.Request.URL), slog.String("body", string(body)))
		return nil, &cModel.APIError{Message: "Downstream service returned an unexpected response", StatusCode: http.StatusInternalServerError}
	}

	return body, nil
}
