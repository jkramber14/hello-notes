package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"

	httptransport "github.com/go-kit/kit/transport/http"
)

const (
	requestIDKey = iota
	pathKey
)

type statusResponse struct {
	Status string `json:"status"`
}

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)

	ctx := context.Background()

	svc := statusEndpoint()

	svc = loggingMiddlware(logger)(svc)

	statusHandler := httptransport.NewServer(
		ctx,
		svc,
		decodeStatusRequest,
		encodeResponse,
		httptransport.ServerBefore(beforeIDExtractor, beforePATHExtractor),
	)

	http.Handle("/", http.FileServer(http.Dir("./build/web")))
	http.Handle("/status", statusHandler)

	port := os.Getenv("PORT")
	logger.Log("listening on", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Log("listen.error", err)
	}

}

func statusEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		return statusResponse{"Testing"}, nil
	}
}

func decodeStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

func beforeIDExtractor(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestIDKey, r.Header.Get("X-Request-Id"))
}

func beforePATHExtractor(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, pathKey, r.URL.EscapedPath())
}

func loggingMiddlware(l log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (result interface{}, err error) {
			var req, resp string

			defer func(b time.Time) {
				l.Log(
					"path", ctx.Value(pathKey),
					"request", req,
					"result", resp,
					"err", err,
					"request_id", ctx.Value(requestIDKey),
					"elapsed", time.Since(b),
				)
			}(time.Now())
			if r, ok := request.(fmt.Stringer); ok {
				req = r.String()
			}
			result, err = next(ctx, request)
			if r, ok := result.(fmt.Stringer); ok {
				resp = r.String()
			}
			return
		}
	}
}
