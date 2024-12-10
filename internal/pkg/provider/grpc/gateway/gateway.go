package gateway

import (
	"github.com/wabee-ai/agenet/pkg/grpc/common"
	gatewaypb "github.com/wabee-ai/agenet/pkg/grpc/gateway"
	"github.com/xgodev/boost/wrapper/log"
)

// gateway implements the GatewayServiceServer interface as a gRPC provider
type gateway struct {
	gatewaypb.UnimplementedGatewayServiceServer
}

// Communicate handles bidirectional streaming for the Gateway
func (p *gateway) Communicate(stream gatewaypb.GatewayService_CommunicateServer) error {
	logger := log.FromContext(stream.Context())

	for {
		// Receive a message from the client
		req, err := stream.Recv()
		if err != nil {
			logger.Errorf("Error receiving message: %v", err)
			return err
		}

		// Log the request details
		logger.Infof("Received request: UUID=%s, RequestType=%s", req.Uuid, req.RequestType)

		// Simulate a response
		response := &common.Response{
			Uuid:        req.Uuid,
			Status:      "SUCCESS",
			ContentType: "application/json",
			Data:        []byte(`{"message": "response from gateway"}`),
		}

		// Send the response back to the client
		if err := stream.Send(response); err != nil {
			logger.Errorf("Error sending response: %v", err)
			return err
		}
	}
}

// NewProvider creates a new instance of the Gateway gRPC gateway
func NewProvider() gatewaypb.GatewayServiceServer {
	return &gateway{}
}
