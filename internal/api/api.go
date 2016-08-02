package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brocaar/loraserver/api"
	"github.com/brocaar/loraserver/internal/api/auth"
	"github.com/brocaar/loraserver/internal/loraserver"
	"github.com/brocaar/loraserver/internal/static"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// GetGRPCServer returns the gRPC API handler.
func GetGRPCServer(ctx context.Context, lsCtx loraserver.Context, validator auth.Validator) *grpc.Server {
	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)

	pb.RegisterApplicationServer(server, NewApplicationAPI(lsCtx, validator))
	pb.RegisterNodeServer(server, NewNodeAPI(lsCtx, validator))
	pb.RegisterChannelListServer(server, NewChannelListAPI(lsCtx, validator))
	pb.RegisterChannelServer(server, NewChannelAPI(lsCtx, validator))
	pb.RegisterNodeSessionServer(server, NewNodeSessionAPI(lsCtx))
	return server
}

// GetJSONGateway returns the JSON gateway for the gRPC API.
func GetJSONGateway(ctx context.Context, lsCtx loraserver.Context, grpcBind string) (http.Handler, error) {
	bindParts := strings.SplitN(grpcBind, ":", 2)
	if len(bindParts) != 2 {
		return nil, errors.New("get port from http-bind failed")
	}
	apiEndpoint := fmt.Sprintf("localhost:%s", bindParts[1])

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	mux := runtime.NewServeMux()

	if err := pb.RegisterApplicationHandlerFromEndpoint(ctx, mux, apiEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register application handler error: %s", err)
	}
	if err := pb.RegisterNodeHandlerFromEndpoint(ctx, mux, apiEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register node handler error: %s", err)
	}
	if err := pb.RegisterChannelListHandlerFromEndpoint(ctx, mux, apiEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register channel-list handler error: %s", err)
	}
	if err := pb.RegisterChannelHandlerFromEndpoint(ctx, mux, apiEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register channel handler error: %s", err)
	}
	if err := pb.RegisterNodeSessionHandlerFromEndpoint(ctx, mux, apiEndpoint, opts); err != nil {
		return nil, fmt.Errorf("register node-session handler error: %s", err)
	}

	return mux, nil
}

// SwaggerHandlerFunc serves the Swagger JSON api documentation.
func SwaggerHandlerFunc(w http.ResponseWriter, r *http.Request) {
	data, err := static.Asset("swagger/index.html")
	if err != nil {
		log.Errorf("get swagger template error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
