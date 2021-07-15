package grpc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/server/config"
)

// StartGRPCWeb starts a gRPC-Web server on the given address.
func StartGRPCWeb(grpcSrv *grpc.Server, config config.Config) (*http.Server, error) {
	var options []grpcweb.Option
	if config.GRPCWeb.EnableUnsafeCORS {
		options = append(options,
			grpcweb.WithOriginFunc(func(origin string) bool {
				return true
			}),
		)
	}

	wrappedServer := grpcweb.WrapServer(grpcSrv, options...)
	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}
	grpcWebSrv := &http.Server{
		Addr:    config.GRPCWeb.Address,
		Handler: http.HandlerFunc(handler),
	}
	errCh := make(chan error)
	go func() {
		if err := grpcWebSrv.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
	}()
	select {
	case err := <-errCh:
		return nil, err
	case <-time.After(5 * time.Second): // assume server started successfully
		return grpcWebSrv, nil
	}
}
