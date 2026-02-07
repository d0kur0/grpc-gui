//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/test.proto

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	"grpc-gui/testserver/proto"
)

type testServer struct {
	proto.UnimplementedTestServiceServer
}

type anotherServer struct {
	proto.UnimplementedAnotherServiceServer
}

func (s *testServer) SimpleCall(ctx context.Context, req *proto.SimpleRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{
		Result:    fmt.Sprintf("Echo: %s", req.Message),
		Processed: req.Value * 2,
	}, nil
}

func (s *testServer) ComplexCall(ctx context.Context, req *proto.ComplexRequest) (*proto.ComplexResponse, error) {
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

func (s *testServer) EmptyCall(ctx context.Context, req *proto.EmptyRequest) (*proto.EmptyResponse, error) {
	return &proto.EmptyResponse{}, nil
}

func (s *testServer) ScheduleTask(ctx context.Context, req *proto.ScheduleRequest) (*proto.ScheduleResponse, error) {
	now := time.Now()
	
	return &proto.ScheduleResponse{
		TaskId:            12345,
		Status:            req.Status,
		CreatedAt:         timestamppb.New(now),
		StartsAt:          req.ScheduledAt,
		EstimatedDuration: req.Timeout,
		AssignedPriority:  req.Priority,
		Message:           fmt.Sprintf("Task '%s' scheduled successfully", req.TaskName),
	}, nil
}

func (s *testServer) ServerStream(req *proto.SimpleRequest, stream proto.TestService_ServerStreamServer) error {
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

func (s *testServer) ClientStream(stream proto.TestService_ClientStreamServer) error {
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

func (s *testServer) BidirectionalStream(stream proto.TestService_BidirectionalStreamServer) error {
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

func (s *anotherServer) GetUser(ctx context.Context, req *proto.SimpleRequest) (*proto.User, error) {
	return &proto.User{
		Id:     1,
		Name:   req.Message,
		Email:  "test@example.com",
		Active: true,
		Status: proto.Status_ACTIVE,
	}, nil
}

func (s *anotherServer) GetUsers(ctx context.Context, req *proto.EmptyRequest) (*proto.ComplexResponse, error) {
	return &proto.ComplexResponse{
		Users: []*proto.User{
			{Id: 1, Name: "User1", Status: proto.Status_ACTIVE},
			{Id: 2, Name: "User2", Status: proto.Status_PENDING},
		},
		Count: 2,
	}, nil
}

func (s *anotherServer) UpdateStatus(ctx context.Context, req *proto.ComplexRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{
		Result:    "Status updated",
		Processed: 1,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	proto.RegisterTestServiceServer(s, &testServer{})
	proto.RegisterAnotherServiceServer(s, &anotherServer{})

	reflection.Register(s)

	fmt.Println("Test gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
