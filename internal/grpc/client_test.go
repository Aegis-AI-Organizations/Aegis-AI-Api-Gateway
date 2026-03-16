package grpc

import (
	"context"
	"log"
	"net"
	"testing"

	v1 "github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/grpc/aegis/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type mockPingServer struct {
	v1.UnimplementedPingServiceServer
}

func (s *mockPingServer) Ping(ctx context.Context, req *v1.PingRequest) (*v1.PingResponse, error) {
	return &v1.PingResponse{Message: "pong"}, nil
}

func TestClient_Ping(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	v1.RegisterPingServiceServer(s, &mockPingServer{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Server exited with error: %v", err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	client := &Client{
		conn:        conn,
		PingService: v1.NewPingServiceClient(conn),
	}

	resp, err := client.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	if resp != "pong" {
		t.Errorf("Expected pong, got %s", resp)
	}
}
