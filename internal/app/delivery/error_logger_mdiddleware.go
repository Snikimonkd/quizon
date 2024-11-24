package delivery

import (
	"context"
	"net/http"

	"quizon/internal/pkg/logger"

	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

func LogErrors(
	f strictnethttp.StrictHTTPHandlerFunc,
	operationID string,
) strictnethttp.StrictHTTPHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
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
