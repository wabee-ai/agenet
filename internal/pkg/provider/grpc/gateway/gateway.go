package gateway

import (
	commonpb "github.com/wabee-ai/agenet/pkg/grpc/common"
	gatewaypb "github.com/wabee-ai/agenet/pkg/grpc/gateway"
	"github.com/xgodev/boost/wrapper/log"
	"time"
)

// gateway implements the GatewayServiceServer interface as a gRPC provider
type gateway struct {
	gatewaypb.UnimplementedGatewayServiceServer
}

// Send handles a unary request and streams multiple responses
func (g *gateway) Send(req *commonpb.Request, stream gatewaypb.GatewayService_SendServer) error {
	logger := log.FromContext(stream.Context())

	// Log the received request
	logger.Infof("Received request: UUID=%s, RequestType=%s", req.Uuid, req.RequestType)

	// Simulate sending multiple responses
	for i := 1; i <= 5; i++ {
		response := &commonpb.Response{
			Uuid:        req.Uuid,
			Status:      "SUCCESS",
			ContentType: "application/json",
			Data:        []byte(`{"message": "response chunk ` + string(i) + `"}`),
		}

		// Stream the response to the client
		if err := stream.Send(response); err != nil {
			logger.Errorf("Error sending response: %v", err)
			return err
		}

		// Simulate delay between responses
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

// New creates a new instance of the Gateway gRPC provider
func New() gatewaypb.GatewayServiceServer {
	return &gateway{}
}
