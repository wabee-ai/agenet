package main

import (
	"context"
	"github.com/wabee-ai/agenet/internal/pkg/provider/grpc/gateway"
	gatewaypb "github.com/wabee-ai/agenet/pkg/grpc/gateway"
	"github.com/xgodev/boost"
	"github.com/xgodev/boost/factory/contrib/google.golang.org/grpc/v1/server"
	"github.com/xgodev/boost/factory/contrib/google.golang.org/grpc/v1/server/plugins/local/wrapper/log"
)

func main() {
	// Create a base context
	ctx := context.Background()

	// Initialize Boost framework
	boost.Start()

	// Create a new gRPC server
	srv := server.NewServer(ctx, log.Register)

	// Register the Gateway
	gatewaypb.RegisterGatewayServiceServer(srv.ServiceRegistrar(), gateway.NewProvider())

	// Start serving
	srv.Serve(ctx)
}
