package main

import (
	"os"
	"testing"

	"grpc-gui/internal/grpcreflect"
	"grpc-gui/internal/models"
	"grpc-gui/internal/testutil"
)

func setupTestApp(t *testing.T) (*App, func()) {
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	app := NewApp(tmpFile.Name())

	err = app.storage.AutoMigrate(&models.Server{}, &models.History{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return app, func() {
		os.Remove(tmpFile.Name())
	}
}

func TestApp_CreateServer(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("Test Server", "localhost:50051", false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	if id == 0 {
		t.Error("expected server ID to be set")
	}
}

func TestApp_CreateServer_WithTLS(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("TLS Server", "localhost:50051", true, true)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	if id == 0 {
		t.Error("expected server ID to be set")
	}

	servers, err := app.GetServers()
	if err != nil {
		t.Fatalf("GetServers failed: %v", err)
	}

	var found *models.Server
	for i := range servers {
		if servers[i].ID == id {
			found = &servers[i]
			break
		}
	}

	if found == nil {
		t.Fatal("server not found")
	}

	if !found.OptUseTLS {
		t.Error("expected OptUseTLS to be true")
	}
	if !found.OptInsecure {
		t.Error("expected OptInsecure to be true")
	}
}

func TestApp_GetServers(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	app.CreateServer("Server 1", "localhost:50051", false, false)
	app.CreateServer("Server 2", "localhost:50052", false, false)

	servers, err := app.GetServers()
	if err != nil {
		t.Fatalf("GetServers failed: %v", err)
	}

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestApp_UpdateServer(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("Original Name", "localhost:50051", false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	err = app.UpdateServer(id, "Updated Name", "localhost:50051", true, true)
	if err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	updated, err := app.storage.GetServer(id)
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", updated.Name)
	}
	if !updated.OptUseTLS {
		t.Error("expected OptUseTLS to be true")
	}
	if !updated.OptInsecure {
		t.Error("expected OptInsecure to be true")
	}
}

func TestApp_DeleteServer(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("To Delete", "localhost:50051", false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	err = app.DeleteServer(id)
	if err != nil {
		t.Fatalf("DeleteServer failed: %v", err)
	}

	_, err = app.storage.GetServer(id)
	if err == nil {
		t.Error("expected error when getting deleted server")
	}
}

func TestApp_GetServerReflection_WithoutTLS(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("Test Server", addr, false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	result, err := app.GetServerReflection(id)
	if err != nil {
		t.Fatalf("GetServerReflection failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected server, got nil")
	}
	if result.Address != addr {
		t.Errorf("expected address %s, got %s", addr, result.Address)
	}
}

func TestApp_GetServerReflection_WithTLS_Insecure(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("TLS Server", addr, true, true)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	result, err := app.GetServerReflection(id)
	if err != nil {
		t.Fatalf("GetServerReflection failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected server, got nil")
	}
	if !result.OptUseTLS {
		t.Error("expected OptUseTLS to be true")
	}
	if !result.OptInsecure {
		t.Error("expected OptInsecure to be true")
	}
}

func TestApp_GetServerReflection_InvalidServer(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	_, err := app.GetServerReflection(999)
	if err == nil {
		t.Error("expected error for invalid server ID")
	}
}

func TestApp_DoGRPCRequest(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("Test Server", addr, false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	payload := `{"message": "test", "value": 42}`
	resp, code, err := app.DoGRPCRequest(id, addr, "testserver.TestService", "SimpleCall", payload, nil, nil)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}

	if code != 0 {
		t.Errorf("expected status code 0, got %d", code)
	}
	if resp == "" {
		t.Error("expected non-empty response")
	}
}

func TestApp_DoGRPCRequest_WithHeaders(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	id, err := app.CreateServer("Test Server", addr, false, false)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	headers := map[string]string{
		"authorization": "Bearer token123",
		"x-custom":      "value",
	}
	contextValues := map[string]string{
		"user-id": "123",
	}

	payload := `{"message": "test"}`
	resp, code, err := app.DoGRPCRequest(id, addr, "testserver.AnotherService", "GetUser", payload, headers, contextValues)
	if err != nil {
		t.Fatalf("DoGRPCRequest failed: %v", err)
	}

	if code != 0 {
		t.Errorf("expected status code 0, got %d", code)
	}
	if resp == "" {
		t.Error("expected non-empty response")
	}
}

func TestApp_GetJsonExample(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	msg := &grpcreflect.MessageInfo{
		Name: "testserver.SimpleRequest",
		Fields: []grpcreflect.FieldInfo{
			{Name: "message", Type: "string"},
			{Name: "value", Type: "int32"},
		},
	}

	json, err := app.GetJsonExample(msg)
	if err != nil {
		t.Fatalf("GetJsonExample failed: %v", err)
	}

	if json == "" {
		t.Error("expected non-empty JSON")
	}
}

func TestApp_ValidateServerAddress_Success(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	result := app.ValidateServerAddress(addr, false, false)
	if result.Status != ValidationStatusSuccess {
		t.Errorf("expected ValidationStatusSuccess, got %d", result.Status)
	}
	if result.Message != "" {
		t.Errorf("expected empty message on success, got %s", result.Message)
	}
}

func TestApp_ValidateServerAddress_ConnectionFailed(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	result := app.ValidateServerAddress("127.0.0.1:1", false, false)
	if result.Status != ValidationStatusConnectionFailed && result.Status != ValidationStatusReflectionNotAvailable {
		t.Errorf("expected ValidationStatusConnectionFailed or ValidationStatusReflectionNotAvailable, got %d", result.Status)
	}
	if result.Message == "" {
		t.Error("expected error message, got empty")
	}
}

func TestApp_ValidateServerAddress_InvalidAddress(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	result := app.ValidateServerAddress("invalid:address:12345", false, false)
	if result.Status != ValidationStatusConnectionFailed && result.Status != ValidationStatusReflectionNotAvailable {
		t.Errorf("expected ValidationStatusConnectionFailed or ValidationStatusReflectionNotAvailable, got %d", result.Status)
	}
	if result.Message == "" {
		t.Error("expected error message, got empty")
	}
}
