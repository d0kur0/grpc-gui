package grpcreflect

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc-gui/internal/utils"
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
	return &proto.SimpleResponse{Result: "test"}, nil
}

func (s *testServer) ComplexCall(ctx context.Context, req *proto.ComplexRequest) (*proto.ComplexResponse, error) {
	return &proto.ComplexResponse{}, nil
}

func (s *testServer) EmptyCall(ctx context.Context, req *proto.EmptyRequest) (*proto.EmptyResponse, error) {
	return &proto.EmptyResponse{}, nil
}

func (s *testServer) ServerStream(req *proto.SimpleRequest, stream proto.TestService_ServerStreamServer) error {
	return nil
}

func (s *testServer) ClientStream(stream proto.TestService_ClientStreamServer) error {
	return nil
}

func (s *testServer) BidirectionalStream(stream proto.TestService_BidirectionalStreamServer) error {
	return nil
}

func (s *anotherServer) GetUser(ctx context.Context, req *proto.SimpleRequest) (*proto.User, error) {
	return &proto.User{}, nil
}

func (s *anotherServer) GetUsers(ctx context.Context, req *proto.EmptyRequest) (*proto.ComplexResponse, error) {
	return &proto.ComplexResponse{}, nil
}

func (s *anotherServer) UpdateStatus(ctx context.Context, req *proto.ComplexRequest) (*proto.SimpleResponse, error) {
	return &proto.SimpleResponse{}, nil
}

func TestNewReflector(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	if reflector.conn == nil {
		t.Error("conn is nil")
	}
	if reflector.client == nil {
		t.Error("client is nil")
	}
}

func TestGetAllServicesInfo(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	if len(servicesInfo.Services) < 2 {
		t.Fatalf("expected at least 2 services, got %d", len(servicesInfo.Services))
	}

	var testService *ServiceInfo
	var anotherService *ServiceInfo

	for i := range servicesInfo.Services {
		if servicesInfo.Services[i].Name == "testserver.TestService" {
			testService = &servicesInfo.Services[i]
		}
		if servicesInfo.Services[i].Name == "testserver.AnotherService" {
			anotherService = &servicesInfo.Services[i]
		}
	}

	if testService == nil {
		t.Fatal("TestService not found")
	}
	if anotherService == nil {
		t.Fatal("AnotherService not found")
	}

	if len(testService.Methods) != 6 {
		t.Errorf("expected 6 methods in TestService, got %d", len(testService.Methods))
	}

	if len(anotherService.Methods) != 3 {
		t.Errorf("expected 3 methods in AnotherService, got %d", len(anotherService.Methods))
	}
}

func TestSimpleRequestResponse(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var simpleCall *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "SimpleCall" {
				simpleCall = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if simpleCall == nil {
		t.Fatal("SimpleCall method not found")
	}

	if simpleCall.RequestType != "testserver.SimpleRequest" {
		t.Errorf("expected RequestType testserver.SimpleRequest, got %s", simpleCall.RequestType)
	}

	if simpleCall.ResponseType != "testserver.SimpleResponse" {
		t.Errorf("expected ResponseType testserver.SimpleResponse, got %s", simpleCall.ResponseType)
	}

	if simpleCall.Request == nil {
		t.Fatal("Request is nil")
	}

	if simpleCall.Response == nil {
		t.Fatal("Response is nil")
	}

	req := simpleCall.Request
	if req.Name != "testserver.SimpleRequest" {
		t.Errorf("expected request name testserver.SimpleRequest, got %s", req.Name)
	}

	if len(req.Fields) != 2 {
		t.Fatalf("expected 2 fields in SimpleRequest, got %d", len(req.Fields))
	}

	var messageField *FieldInfo
	var valueField *FieldInfo

	for i := range req.Fields {
		switch req.Fields[i].Name {
		case "message":
			messageField = &req.Fields[i]
		case "value":
			valueField = &req.Fields[i]
		}
	}

	if messageField == nil {
		t.Fatal("message field not found")
	}
	if messageField.Type != "string" {
		t.Errorf("expected message field type string, got %s", messageField.Type)
	}
	if messageField.Number != 1 {
		t.Errorf("expected message field number 1, got %d", messageField.Number)
	}

	if valueField == nil {
		t.Fatal("value field not found")
	}
	if valueField.Type != "int32" {
		t.Errorf("expected value field type int32, got %s", valueField.Type)
	}
	if valueField.Number != 2 {
		t.Errorf("expected value field number 2, got %d", valueField.Number)
	}
}

func TestUserMessageFields(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var getUserMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "GetUser" {
				getUserMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if getUserMethod == nil {
		t.Fatal("GetUser method not found")
	}

	userMsg := getUserMethod.Response
	if userMsg == nil {
		t.Fatal("Response is nil")
	}

	if userMsg.Name != "testserver.User" {
		t.Errorf("expected User message, got %s", userMsg.Name)
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range userMsg.Fields {
		fieldMap[userMsg.Fields[i].Name] = &userMsg.Fields[i]
	}

	tests := []struct {
		name     string
		typeStr  string
		number   int32
		repeated bool
		optional bool
	}{
		{"id", "int64", 1, false, true},
		{"name", "string", 2, false, true},
		{"email", "string", 3, false, true},
		{"active", "bool", 4, false, true},
		{"balance", "double", 5, false, true},
		{"score", "float", 6, false, true},
		{"age", "int32", 7, false, true},
		{"points", "uint32", 8, false, true},
		{"signed_points", "sint32", 9, false, true},
		{"fixed_id", "fixed64", 10, false, true},
		{"sfixed_value", "sfixed32", 11, false, true},
		{"avatar", "bytes", 12, false, true},
		{"status", "testserver.Status", 13, false, true},
		{"tags", "string", 14, true, false},
		{"numbers", "int32", 15, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, ok := fieldMap[tt.name]
			if !ok {
				t.Fatalf("field %s not found", tt.name)
			}

			if field.Type != tt.typeStr {
				t.Errorf("expected type %s, got %s", tt.typeStr, field.Type)
			}

			if field.Number != tt.number {
				t.Errorf("expected number %d, got %d", tt.number, field.Number)
			}

			if field.Repeated != tt.repeated {
				t.Errorf("expected repeated %v, got %v", tt.repeated, field.Repeated)
			}

			if field.Optional != tt.optional {
				t.Errorf("expected optional %v, got %v", tt.optional, field.Optional)
			}

			if tt.name == "status" {
				if !field.IsEnum {
					t.Error("status field should be marked as enum")
				}
				if len(field.EnumValues) == 0 {
					t.Error("status field should have enum values")
				}
			} else {
				if field.IsEnum {
					t.Errorf("field %s should not be marked as enum", tt.name)
				}
			}
		})
	}

	addressField, ok := fieldMap["address"]
	if !ok {
		t.Fatal("address field not found")
	}

	if addressField.Type != "testserver.Address" {
		t.Errorf("expected address type testserver.Address, got %s", addressField.Type)
	}

	if addressField.Message == nil {
		t.Fatal("address nested message is nil")
	}

	if addressField.Message.Name != "testserver.Address" {
		t.Errorf("expected nested message name testserver.Address, got %s", addressField.Message.Name)
	}

	addressesField, ok := fieldMap["addresses"]
	if !ok {
		t.Fatal("addresses field not found")
	}

	if !addressesField.Repeated {
		t.Error("addresses field should be repeated")
	}

	if addressesField.Message == nil {
		t.Fatal("addresses nested message is nil")
	}
}

func TestMapFields(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	userMapField, ok := fieldMap["user_map"]
	if !ok {
		t.Fatal("user_map field not found")
	}

	if !userMapField.IsMap {
		t.Error("user_map field should be a map")
	}

	if userMapField.MapKey != "string" {
		t.Errorf("expected map key type string, got %s", userMapField.MapKey)
	}

	if userMapField.MapValue != "testserver.User" {
		t.Errorf("expected map value type testserver.User, got %s", userMapField.MapValue)
	}

	if userMapField.Type != "map<string, testserver.User>" {
		t.Errorf("expected type map<string, testserver.User>, got %s", userMapField.Type)
	}

	if userMapField.Repeated {
		t.Error("map field should not be marked as repeated")
	}

	if userMapField.Message == nil {
		t.Fatal("user_map nested message is nil")
	}

	statusMapField, ok := fieldMap["status_map"]
	if !ok {
		t.Fatal("status_map field not found")
	}

	if !statusMapField.IsMap {
		t.Error("status_map field should be a map")
	}

	if statusMapField.MapKey != "string" {
		t.Errorf("expected map key type string, got %s", statusMapField.MapKey)
	}

	if statusMapField.MapValue != "testserver.Status" {
		t.Errorf("expected map value type testserver.Status, got %s", statusMapField.MapValue)
	}

	idToNameField, ok := fieldMap["id_to_name"]
	if !ok {
		t.Fatal("id_to_name field not found")
	}

	if !idToNameField.IsMap {
		t.Error("id_to_name field should be a map")
	}

	if idToNameField.MapKey != "int64" {
		t.Errorf("expected map key type int64, got %s", idToNameField.MapKey)
	}

	if idToNameField.MapValue != "string" {
		t.Errorf("expected map value type string, got %s", idToNameField.MapValue)
	}

	userField, ok := fieldMap["user"]
	if !ok {
		t.Fatal("user field not found")
	}

	if userField.Message == nil {
		t.Fatal("user nested message is nil")
	}

	userMsg := userField.Message
	userFieldMap := make(map[string]*FieldInfo)
	for i := range userMsg.Fields {
		userFieldMap[userMsg.Fields[i].Name] = &userMsg.Fields[i]
	}

	metadataField, ok := userFieldMap["metadata"]
	if !ok {
		t.Fatal("metadata field not found in User")
	}

	if !metadataField.IsMap {
		t.Error("metadata field should be a map")
	}

	if metadataField.MapKey != "string" {
		t.Errorf("expected map key type string, got %s", metadataField.MapKey)
	}

	if metadataField.MapValue != "string" {
		t.Errorf("expected map value type string, got %s", metadataField.MapValue)
	}

	addressMapField, ok := userFieldMap["address_map"]
	if !ok {
		t.Fatal("address_map field not found in User")
	}

	if !addressMapField.IsMap {
		t.Error("address_map field should be a map")
	}

	if addressMapField.MapKey != "int32" {
		t.Errorf("expected map key type int32, got %s", addressMapField.MapKey)
	}

	if addressMapField.MapValue != "testserver.Address" {
		t.Errorf("expected map value type testserver.Address, got %s", addressMapField.MapValue)
	}

	if addressMapField.Message == nil {
		t.Fatal("address_map nested message is nil")
	}
}

func TestNestedMessage(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	nestedField, ok := fieldMap["nested"]
	if !ok {
		t.Fatal("nested field not found")
	}

	if nestedField.Message == nil {
		t.Fatal("nested message is nil")
	}

	nestedMsg := nestedField.Message
	if nestedMsg.Name != "testserver.NestedMessage" {
		t.Errorf("expected NestedMessage, got %s", nestedMsg.Name)
	}

	nestedFieldMap := make(map[string]*FieldInfo)
	for i := range nestedMsg.Fields {
		nestedFieldMap[nestedMsg.Fields[i].Name] = &nestedMsg.Fields[i]
	}

	nestedNestedField, ok := nestedFieldMap["nested"]
	if !ok {
		t.Fatal("nested.nested field not found")
	}

	if nestedNestedField.Message == nil {
		t.Fatal("nested.nested message is nil")
	}

	if nestedNestedField.Message.Name != "testserver.NestedMessage" {
		t.Errorf("expected recursive NestedMessage, got %s", nestedNestedField.Message.Name)
	}

	childrenField, ok := nestedFieldMap["children"]
	if !ok {
		t.Fatal("children field not found")
	}

	if !childrenField.Repeated {
		t.Error("children field should be repeated")
	}

	if childrenField.Message == nil {
		t.Fatal("children nested message is nil")
	}
}

func TestEnumField(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	statusField, ok := fieldMap["status"]
	if !ok {
		t.Fatal("status field not found")
	}

	if statusField.Type != "testserver.Status" {
		t.Errorf("expected enum type testserver.Status, got %s", statusField.Type)
	}

	if !statusField.IsEnum {
		t.Error("status field should be marked as enum")
	}

	if !statusField.IsEnum {
		t.Error("status field should be marked as enum")
	}

	statusesField, ok := fieldMap["statuses"]
	if !ok {
		t.Fatal("statuses field not found")
	}

	if !statusesField.Repeated {
		t.Error("statuses field should be repeated")
	}

	if statusesField.Type != "testserver.Status" {
		t.Errorf("expected enum type testserver.Status, got %s", statusesField.Type)
	}

	if !statusesField.IsEnum {
		t.Error("statuses field should be marked as enum")
	}

	if len(statusField.EnumValues) == 0 {
		t.Error("expected enum values to be populated")
	}

	expectedEnumValues := map[string]int32{
		"UNKNOWN":  0,
		"PENDING":  1,
		"ACTIVE":   2,
		"INACTIVE": 3,
		"DELETED":  4,
	}

	if len(statusField.EnumValues) != len(expectedEnumValues) {
		t.Errorf("expected %d enum values, got %d", len(expectedEnumValues), len(statusField.EnumValues))
	}

	for _, enumValue := range statusField.EnumValues {
		expectedNumber, ok := expectedEnumValues[enumValue.Name]
		if !ok {
			t.Errorf("unexpected enum value: %s", enumValue.Name)
			continue
		}
		if enumValue.Number != expectedNumber {
			t.Errorf("enum value %s expected number %d, got %d", enumValue.Name, expectedNumber, enumValue.Number)
		}
	}
}

func TestIsEnumField(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	enumFields := []string{"status", "statuses"}
	nonEnumFields := []string{"user", "users", "timestamps", "prices", "flags", "raw_data"}

	for _, fieldName := range enumFields {
		field, ok := fieldMap[fieldName]
		if !ok {
			t.Errorf("field %s not found", fieldName)
			continue
		}

		if !field.IsEnum {
			t.Errorf("field %s should be marked as enum", fieldName)
		}

		if len(field.EnumValues) == 0 {
			t.Errorf("field %s should have enum values", fieldName)
		}
	}

	for _, fieldName := range nonEnumFields {
		field, ok := fieldMap[fieldName]
		if !ok {
			continue
		}

		if field.IsEnum {
			t.Errorf("field %s should not be marked as enum", fieldName)
		}
	}

	statusMapField, ok := fieldMap["status_map"]
	if ok {
		if statusMapField.IsEnum {
			t.Error("status_map field should not be marked as enum (it's a map with enum values)")
		}

		if !statusMapField.IsMap {
			t.Error("status_map field should be marked as map")
		}
	}
}

func TestClose(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}

	err = reflector.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestEmptyMessage(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var emptyCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "EmptyCall" {
				emptyCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if emptyCallMethod == nil {
		t.Fatal("EmptyCall method not found")
	}

	if emptyCallMethod.RequestType != "testserver.EmptyRequest" {
		t.Errorf("expected RequestType testserver.EmptyRequest, got %s", emptyCallMethod.RequestType)
	}

	if emptyCallMethod.ResponseType != "testserver.EmptyResponse" {
		t.Errorf("expected ResponseType testserver.EmptyResponse, got %s", emptyCallMethod.ResponseType)
	}

	if emptyCallMethod.Request != nil && len(emptyCallMethod.Request.Fields) != 0 {
		t.Errorf("expected empty request fields, got %d", len(emptyCallMethod.Request.Fields))
	}

	if emptyCallMethod.Response != nil && len(emptyCallMethod.Response.Fields) != 0 {
		t.Errorf("expected empty response fields, got %d", len(emptyCallMethod.Response.Fields))
	}
}

func TestRepeatedFields(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	timestampsField, ok := fieldMap["timestamps"]
	if !ok {
		t.Fatal("timestamps field not found")
	}

	if !timestampsField.Repeated {
		t.Error("timestamps field should be repeated")
	}

	if timestampsField.Type != "int64" {
		t.Errorf("expected type int64, got %s", timestampsField.Type)
	}

	pricesField, ok := fieldMap["prices"]
	if !ok {
		t.Fatal("prices field not found")
	}

	if !pricesField.Repeated {
		t.Error("prices field should be repeated")
	}

	if pricesField.Type != "double" {
		t.Errorf("expected type double, got %s", pricesField.Type)
	}

	flagsField, ok := fieldMap["flags"]
	if !ok {
		t.Fatal("flags field not found")
	}

	if !flagsField.Repeated {
		t.Error("flags field should be repeated")
	}

	if flagsField.Type != "bool" {
		t.Errorf("expected type bool, got %s", flagsField.Type)
	}

	usersField, ok := fieldMap["users"]
	if !ok {
		t.Fatal("users field not found")
	}

	if !usersField.Repeated {
		t.Error("users field should be repeated")
	}

	if usersField.Message == nil {
		t.Fatal("users nested message is nil")
	}

	if usersField.Message.Name != "testserver.User" {
		t.Errorf("expected nested message testserver.User, got %s", usersField.Message.Name)
	}
}

func TestFieldNumbers(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	if req == nil {
		t.Fatal("Request is nil")
	}

	fieldByNumber := make(map[int32]*FieldInfo)
	for i := range req.Fields {
		fieldByNumber[req.Fields[i].Number] = &req.Fields[i]
	}

	expectedFields := map[int32]string{
		1:  "user",
		2:  "users",
		3:  "user_map",
		4:  "nested",
		5:  "status",
		6:  "statuses",
		7:  "status_map",
		11: "timestamps",
		12: "prices",
		13: "flags",
		14: "id_to_name",
		15: "raw_data",
	}

	for num, name := range expectedFields {
		field, ok := fieldByNumber[num]
		if !ok {
			t.Errorf("field number %d (%s) not found", num, name)
			continue
		}

		if field.Name != name {
			t.Errorf("field number %d expected name %s, got %s", num, name, field.Name)
		}
	}
}

func TestRequestExampleAndSchema(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	if len(complexCallMethod.RequestExample) == 0 {
		t.Error("RequestExample should not be empty")
	}

	if len(complexCallMethod.RequestSchema) == 0 {
		t.Error("RequestSchema should not be empty")
	}

	if len(complexCallMethod.ResponseExample) == 0 {
		t.Error("ResponseExample should not be empty")
	}

	var requestExample map[string]interface{}
	if err := json.Unmarshal(complexCallMethod.RequestExample, &requestExample); err != nil {
		t.Fatalf("Failed to unmarshal RequestExample: %v", err)
	}

	if _, ok := requestExample["status"]; !ok {
		t.Error("RequestExample should contain status field")
	}

	var requestSchema map[string]interface{}
	if err := json.Unmarshal(complexCallMethod.RequestSchema, &requestSchema); err != nil {
		t.Fatalf("Failed to unmarshal RequestSchema: %v", err)
	}

	statusField, ok := requestSchema["status"].(map[string]interface{})
	if !ok {
		t.Fatal("status field should be an object in RequestSchema")
	}

	if isEnum, ok := statusField["isEnum"].(bool); !ok || !isEnum {
		t.Error("status field should have isEnum: true")
	}

	enumValues, ok := statusField["enumValues"].([]interface{})
	if !ok {
		t.Fatal("status field should have enumValues array")
	}

	expectedEnumValues := []string{"UNKNOWN", "PENDING", "ACTIVE", "INACTIVE", "DELETED"}
	if len(enumValues) != len(expectedEnumValues) {
		t.Errorf("expected %d enum values, got %d", len(expectedEnumValues), len(enumValues))
	}

	for i, expected := range expectedEnumValues {
		if i >= len(enumValues) {
			break
		}
		if enumValues[i] != expected {
			t.Errorf("enum value at index %d expected %s, got %v", i, expected, enumValues[i])
		}
	}
}

func TestOneofGroup(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var complexCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "ComplexCall" {
				complexCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if complexCallMethod == nil {
		t.Fatal("ComplexCall method not found")
	}

	req := complexCallMethod.Request
	fieldMap := make(map[string]*FieldInfo)
	for i := range req.Fields {
		fieldMap[req.Fields[i].Name] = &req.Fields[i]
	}

	oneofFields := []string{"text", "number", "user_payload"}
	for _, fieldName := range oneofFields {
		field, ok := fieldMap[fieldName]
		if !ok {
			t.Errorf("oneof field %s not found", fieldName)
			continue
		}

		if field.OneofGroup != "payload" {
			t.Errorf("field %s expected oneofGroup 'payload', got '%s'", fieldName, field.OneofGroup)
		}
	}

	var requestSchema map[string]interface{}
	if err := json.Unmarshal(complexCallMethod.RequestSchema, &requestSchema); err != nil {
		t.Fatalf("Failed to unmarshal RequestSchema: %v", err)
	}

	for _, fieldName := range oneofFields {
		fieldSchema, ok := requestSchema[fieldName].(map[string]interface{})
		if !ok {
			t.Errorf("field %s should be an object in RequestSchema", fieldName)
			continue
		}

		oneofGroup, ok := fieldSchema["oneofGroup"].(string)
		if !ok || oneofGroup != "payload" {
			t.Errorf("field %s expected oneofGroup 'payload' in schema, got %v", fieldName, fieldSchema["oneofGroup"])
		}
	}
}

func TestIsReflectionService(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"grpc.reflection.v1.ServerReflection", true},
		{"grpc.reflection.v1alpha.ServerReflection", true},
		{"grpc.reflection.v1beta.ServerReflection", true},
		{"testserver.TestService", false},
		{"testserver.AnotherService", false},
		{"some.other.Service", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReflectionService(tt.name)
			if result != tt.expected {
				t.Errorf("IsReflectionService(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}
