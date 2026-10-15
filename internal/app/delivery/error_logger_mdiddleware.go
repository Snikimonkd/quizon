package delivery

import (
	"context"
	"net/http"

	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"

	"quizon/internal/pkg/logger"
)

func LogErrors(
	f strictnethttp.StrictHTTPHandlerFunc,
	operationID string,
) strictnethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		resp, err := f(ctx, w, r, request)
		if err != nil {
			logger.Errorf(
				"operation: %v, url: %v, request: %v, error: %v",
				operationID,
				r.URL.RawPath,
				request,
				err.Error(),
			)
		}
		return resp, nil
	}
}
