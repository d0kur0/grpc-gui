package main

import (
	"context"
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
	app.startup(context.Background())

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

	server := &models.Server{
		Name:        "Test Server",
		Address:     "localhost:50051",
		OptUseTLS:   false,
		OptInsecure: false,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	if server.ID == 0 {
		t.Error("expected server ID to be set")
	}
}

func TestApp_CreateServer_WithTLS(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	server := &models.Server{
		Name:        "TLS Server",
		Address:     "localhost:50051",
		OptUseTLS:   true,
		OptInsecure: true,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	if !server.OptUseTLS {
		t.Error("expected OptUseTLS to be true")
	}
	if !server.OptInsecure {
		t.Error("expected OptInsecure to be true")
	}
}

func TestApp_GetServers(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	server1 := &models.Server{
		Name:    "Server 1",
		Address: "localhost:50051",
	}
	server2 := &models.Server{
		Name:    "Server 2",
		Address: "localhost:50052",
	}

	app.CreateServer(server1)
	app.CreateServer(server2)

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

	server := &models.Server{
		Name:        "Original Name",
		Address:     "localhost:50051",
		OptUseTLS:   false,
		OptInsecure: false,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	server.Name = "Updated Name"
	server.OptUseTLS = true
	server.OptInsecure = true

	err = app.UpdateServer(server)
	if err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	updated, err := app.storage.GetServer(server.ID)
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

	server := &models.Server{
		Name:    "To Delete",
		Address: "localhost:50051",
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	err = app.DeleteServer(server.ID)
	if err != nil {
		t.Fatalf("DeleteServer failed: %v", err)
	}

	_, err = app.storage.GetServer(server.ID)
	if err == nil {
		t.Error("expected error when getting deleted server")
	}
}

func TestApp_GetServerReflection_WithoutTLS(t *testing.T) {
	addr, stop := testutil.StartTestServer(t)
	defer stop()

	app, cleanup := setupTestApp(t)
	defer cleanup()

	server := &models.Server{
		Name:        "Test Server",
		Address:     addr,
		OptUseTLS:   false,
		OptInsecure: false,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	result, err := app.GetServerReflection(server.ID)
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

	server := &models.Server{
		Name:        "TLS Server",
		Address:     addr,
		OptUseTLS:   true,
		OptInsecure: true,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	result, err := app.GetServerReflection(server.ID)
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

	server := &models.Server{
		Name:    "Test Server",
		Address: addr,
	}

	err := app.CreateServer(server)
	if err != nil {
		t.Fatalf("CreateServer failed: %v", err)
	}

	payload := `{"message": "test", "value": 42}`
	resp, code, err := app.DoGRPCRequest(server.ID, addr, "testserver.TestService", "SimpleCall", payload, nil, nil)
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

	server := &models.Server{
		Name:    "Test Server",
		Address: addr,
	}

	err := app.CreateServer(server)
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
	resp, code, err := app.DoGRPCRequest(server.ID, addr, "testserver.AnotherService", "GetUser", payload, headers, contextValues)
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
