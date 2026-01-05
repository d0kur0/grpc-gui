package grpcrequest

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-gui/testserver/proto"
)

func startTestServer(t *testing.T) (string, func()) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	addr := lis.Addr().String()

	s := grpc.NewServer()
	proto.RegisterTestServiceServer(s, &testServer{})
	proto.RegisterAnotherServiceServer(s, &anotherServer{})
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

type testServer struct {
	proto.UnimplementedTestServiceServer
}

type anotherServer struct {
	proto.UnimplementedAnotherServiceServer
}

func (s *testServer) SimpleCall(ctx context.Context, req *proto.SimpleRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{
		Result:    "Echo: " + req.Message,
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

func TestDoGRPCRequest_SimpleCall(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	payload := `{"message": "test", "value": 42}`

	resp, code, _, err := DoGRPCRequest(addr, "testserver.TestService", "SimpleCall", payload, nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected code 0 (OK), got %d", code)
	}

	var result proto.SimpleResponse
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.Result != "Echo: test" {
		t.Errorf("expected result 'Echo: test', got '%s'", result.Result)
	}
	if result.Processed != 84 {
		t.Errorf("expected processed 84, got %d", result.Processed)
	}
}

func TestDoGRPCRequest_EmptyCall(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	resp, code, _, err := DoGRPCRequest(addr, "testserver.TestService", "EmptyCall", "", nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected code 0 (OK), got %d", code)
	}

	var result proto.EmptyResponse
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
}

func TestDoGRPCRequest_ComplexCall(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	payload := `{
		"user": {
			"id": 1,
			"name": "Test User",
			"email": "test@example.com",
			"active": true,
			"status": "ACTIVE"
		},
		"status": "ACTIVE"
	}`

	resp, code, _, err := DoGRPCRequest(addr, "testserver.TestService", "ComplexCall", payload, nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected code 0 (OK), got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count, ok := result["count"].(float64); !ok || int(count) != 1 {
		t.Errorf("expected count 1, got %v", result["count"])
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok || user["name"] != "Test User" {
		t.Errorf("expected user name 'Test User', got '%v'", result["user"])
	}

	if total, ok := result["total"].(float64); !ok || total != 100.5 {
		t.Errorf("expected total 100.5, got %v", result["total"])
	}
}

func TestDoGRPCRequest_GetUser(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	payload := `{"message": "John Doe"}`

	resp, code, _, err := DoGRPCRequest(addr, "testserver.AnotherService", "GetUser", payload, nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected code 0 (OK), got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if id, ok := result["id"].(string); !ok || id != "1" {
		t.Errorf("expected id '1', got %v", result["id"])
	}
	if result["name"] != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%v'", result["name"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%v'", result["email"])
	}
	if result["active"] != true {
		t.Error("expected active true, got false")
	}
}

func TestDoGRPCRequest_GetUsers(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	resp, code, _, err := DoGRPCRequest(addr, "testserver.AnotherService", "GetUsers", "", nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}
	if code != 0 {
		t.Errorf("expected code 0 (OK), got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if count, ok := result["count"].(float64); !ok || int(count) != 2 {
		t.Errorf("expected count 2, got %v", result["count"])
	}

	users, ok := result["users"].([]interface{})
	if !ok || len(users) != 2 {
		t.Errorf("expected 2 users, got %v", result["users"])
	}

	if len(users) > 0 {
		user1, ok := users[0].(map[string]interface{})
		if !ok || user1["name"] != "User1" {
			t.Errorf("expected first user name 'User1', got '%v'", users[0])
		}
	}
}

func TestDoGRPCRequest_InvalidService(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	_, code, _, err := DoGRPCRequest(addr, "testserver.InvalidService", "SimpleCall", "", nil, nil)
	if err == nil {
		t.Error("expected error for invalid service, got nil")
	}
	if code == 0 {
		t.Error("expected non-zero status code for invalid service")
	}
}

func TestDoGRPCRequest_InvalidMethod(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	_, code, _, err := DoGRPCRequest(addr, "testserver.TestService", "InvalidMethod", "", nil, nil)
	if err == nil {
		t.Error("expected error for invalid method, got nil")
	}
	if code == 0 {
		t.Error("expected non-zero status code for invalid method")
	}
}

func TestDoGRPCRequest_InvalidPayload(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	_, code, _, err := DoGRPCRequest(addr, "testserver.TestService", "SimpleCall", "invalid json", nil, nil)
	if err == nil {
		t.Error("expected error for invalid payload, got nil")
	}
	if code == 0 {
		t.Error("expected non-zero status code for invalid payload")
	}
}
