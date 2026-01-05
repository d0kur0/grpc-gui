package testutil

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-gui/testserver/proto"
)

func StartTestServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	addr := lis.Addr().String()

	s := grpc.NewServer()
	proto.RegisterTestServiceServer(s, &TestServer{})
	proto.RegisterAnotherServiceServer(s, &AnotherServer{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return addr, func() {
		s.Stop()
		lis.Close()
	}
}

type TestServer struct {
	proto.UnimplementedTestServiceServer
}

type AnotherServer struct {
	proto.UnimplementedAnotherServiceServer
}

func (s *TestServer) SimpleCall(ctx context.Context, req *proto.SimpleRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{
		Result:    fmt.Sprintf("Echo: %s", req.Message),
		Processed: req.Value * 2,
	}, nil
}

func (s *TestServer) ComplexCall(ctx context.Context, req *proto.ComplexRequest) (*proto.ComplexResponse, error) {
	users := []*proto.User{req.User}
	if len(req.Users) > 0 {
		users = append(users, req.Users...)
	}

	return &proto.ComplexResponse{
		User:     req.User,
		Users:    users,
		Status:   req.Status,
		Count:    int32(len(users)),
		Total:    100.5,
		Nested:   req.Nested,
		Messages: []string{"success", "processed"},
	}, nil
}

func (s *TestServer) EmptyCall(ctx context.Context, req *proto.EmptyRequest) (*proto.EmptyResponse, error) {
	return &proto.EmptyResponse{}, nil
}

func (s *TestServer) ServerStream(req *proto.SimpleRequest, stream proto.TestService_ServerStreamServer) error {
	for i := 0; i < 5; i++ {
		if err := stream.Send(&proto.StreamResponse{
			Id:     int32(i),
			Result: fmt.Sprintf("%s-%d", req.Message, i),
			Status: proto.Status_ACTIVE,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *TestServer) ClientStream(stream proto.TestService_ClientStreamServer) error {
	count := 0
	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}
		count++
		_ = req
	}

	return stream.SendAndClose(&proto.ComplexResponse{
		Count:  int32(count),
		Status: proto.Status_ACTIVE,
	})
}

func (s *TestServer) BidirectionalStream(stream proto.TestService_BidirectionalStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			break
		}

		if err := stream.Send(&proto.StreamResponse{
			Id:     req.Id,
			Result: fmt.Sprintf("Echo: %s", req.Data),
			Status: proto.Status_ACTIVE,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *AnotherServer) GetUser(ctx context.Context, req *proto.SimpleRequest) (*proto.User, error) {
	return &proto.User{
		Id:     1,
		Name:   req.Message,
		Email:  "test@example.com",
		Active: true,
		Status: proto.Status_ACTIVE,
	}, nil
}

func (s *AnotherServer) GetUsers(ctx context.Context, req *proto.EmptyRequest) (*proto.ComplexResponse, error) {
	return &proto.ComplexResponse{
		Users: []*proto.User{
			{Id: 1, Name: "User1", Status: proto.Status_ACTIVE},
			{Id: 2, Name: "User2", Status: proto.Status_PENDING},
		},
		Count: 2,
	}, nil
}

func (s *AnotherServer) UpdateStatus(ctx context.Context, req *proto.ComplexRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{
		Result:    "Status updated",
		Processed: 1,
	}, nil
}
