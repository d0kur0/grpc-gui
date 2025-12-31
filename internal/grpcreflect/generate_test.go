package grpcreflect

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGenerateJSONExample_SimpleRequest(t *testing.T) {
	msg := &MessageInfo{
		Name: "testserver.SimpleRequest",
		Fields: []FieldInfo{
			{Name: "message", Type: "string", Number: 1},
			{Name: "value", Type: "int32", Number: 2},
		},
	}

	jsonBytes, err := GenerateJSONExample(msg)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	t.Logf("Generated JSON:\n%s", string(jsonBytes))

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["message"] != "" {
		t.Errorf("expected empty string for message, got %v", result["message"])
	}

	if result["value"] != float64(0) {
		t.Errorf("expected 0 for value, got %v", result["value"])
	}
}

func TestGenerateJSONExample_ComplexRequest(t *testing.T) {
	addressMsg := &MessageInfo{
		Name: "testserver.Address",
		Fields: []FieldInfo{
			{Name: "street", Type: "string", Number: 1},
			{Name: "city", Type: "string", Number: 2},
			{Name: "zip_code", Type: "int32", Number: 4},
		},
	}

	userMsg := &MessageInfo{
		Name: "testserver.User",
		Fields: []FieldInfo{
			{Name: "id", Type: "int64", Number: 1},
			{Name: "name", Type: "string", Number: 2},
			{Name: "active", Type: "bool", Number: 4},
			{Name: "balance", Type: "double", Number: 5},
			{Name: "address", Type: "testserver.Address", Number: 16, Message: addressMsg},
		},
	}

	msg := &MessageInfo{
		Name: "testserver.ComplexRequest",
		Fields: []FieldInfo{
			{Name: "user", Type: "testserver.User", Number: 1, Message: userMsg},
			{Name: "status", Type: "testserver.Status", Number: 5, EnumValues: []EnumValueInfo{
				{Name: "UNKNOWN", Number: 0},
				{Name: "PENDING", Number: 1},
				{Name: "ACTIVE", Number: 2},
			}},
			{Name: "timestamps", Type: "int64", Number: 11, Repeated: true},
		},
	}

	jsonBytes, err := GenerateJSONExample(msg)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user should be an object")
	}

	if user["name"] != "" {
		t.Errorf("expected empty string for user.name, got %v", user["name"])
	}

	if user["active"] != false {
		t.Errorf("expected false for user.active, got %v", user["active"])
	}

	address, ok := user["address"].(map[string]interface{})
	if !ok {
		t.Fatal("user.address should be an object")
	}

	if address["street"] != "" {
		t.Errorf("expected empty string for address.street, got %v", address["street"])
	}

	timestamps, ok := result["timestamps"].([]interface{})
	if !ok {
		t.Fatal("timestamps should be an array")
	}

	if len(timestamps) != 1 {
		t.Errorf("expected array with 1 element, got %d", len(timestamps))
	}

	status, ok := result["status"].(string)
	if !ok {
		t.Fatal("status should be a string")
	}

	if status != "UNKNOWN" {
		t.Errorf("expected status to be UNKNOWN (first enum value), got %s", status)
	}
}

func TestGenerateJSONExample_WithMap(t *testing.T) {
	msg := &MessageInfo{
		Name: "testserver.ComplexRequest",
		Fields: []FieldInfo{
			{Name: "user_map", Type: "map<string, testserver.User>", Number: 3, IsMap: true, MapKey: "string", MapValue: "testserver.User"},
			{Name: "metadata", Type: "map<string, string>", Number: 17, IsMap: true, MapKey: "string", MapValue: "string"},
		},
	}

	jsonBytes, err := GenerateJSONExample(msg)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	userMap, ok := result["user_map"].(map[string]interface{})
	if !ok {
		t.Fatal("user_map should be an object")
	}

	if len(userMap) != 0 {
		t.Errorf("expected empty map, got %d elements", len(userMap))
	}

	metadata, ok := result["metadata"].(map[string]interface{})
	if !ok {
		t.Fatal("metadata should be an object")
	}

	if len(metadata) != 0 {
		t.Errorf("expected empty map, got %d elements", len(metadata))
	}
}

func TestGenerateJSONExample_EmptyMessage(t *testing.T) {
	msg := &MessageInfo{
		Name:   "testserver.EmptyRequest",
		Fields: []FieldInfo{},
	}

	jsonBytes, err := GenerateJSONExample(msg)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty object, got %d fields", len(result))
	}
}

func TestGenerateJSONExample_NilMessage(t *testing.T) {
	jsonBytes, err := GenerateJSONExample(nil)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	if string(jsonBytes) != "{}" {
		t.Errorf("expected empty JSON object, got %s", string(jsonBytes))
	}
}

func TestGenerateJSONExample_Integration(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, ReflectorOptions{UseTLS: false})
	if err != nil {
		t.Fatalf("NewReflector failed: %v", err)
	}
	defer reflector.Close()

	servicesInfo, err := reflector.GetAllServicesInfo()
	if err != nil {
		t.Fatalf("GetAllServicesInfo failed: %v", err)
	}

	var simpleCallMethod *MethodInfo
	for i := range servicesInfo.Services {
		for j := range servicesInfo.Services[i].Methods {
			if servicesInfo.Services[i].Methods[j].Name == "SimpleCall" {
				simpleCallMethod = &servicesInfo.Services[i].Methods[j]
				break
			}
		}
	}

	if simpleCallMethod == nil {
		t.Fatal("SimpleCall method not found")
	}

	jsonBytes, err := GenerateJSONExample(simpleCallMethod.Request)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["message"] != "" {
		t.Errorf("expected empty string for message, got %v", result["message"])
	}

	if result["value"] != float64(0) {
		t.Errorf("expected 0 for value, got %v", result["value"])
	}
}

func TestGenerateJSONExample_ComplexIntegration(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx := context.Background()
	reflector, err := NewReflector(ctx, addr, ReflectorOptions{UseTLS: false})
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

	jsonBytes, err := GenerateJSONExample(complexCallMethod.Request)
	if err != nil {
		t.Fatalf("GenerateJSONExample failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected non-empty result")
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user should be an object")
	}

	if len(user) == 0 {
		t.Error("user object should not be empty")
	}

	users, ok := result["users"].([]interface{})
	if !ok {
		t.Fatal("users should be an array")
	}

	if len(users) != 1 {
		t.Errorf("expected array with 1 element, got %d", len(users))
	}

	userMap, ok := result["user_map"].(map[string]interface{})
	if !ok {
		t.Fatal("user_map should be an object")
	}

	if len(userMap) != 0 {
		t.Errorf("expected empty map, got %d elements", len(userMap))
	}

	statusField := findField(complexCallMethod.Request, "status")
	if statusField != nil && len(statusField.EnumValues) > 0 {
		status, ok := result["status"].(string)
		if ok {
			if status != statusField.EnumValues[0].Name {
				t.Errorf("expected status to be first enum value %s, got %s", statusField.EnumValues[0].Name, status)
			}
		}
	}
}

func findField(msg *MessageInfo, fieldName string) *FieldInfo {
	if msg == nil {
		return nil
	}
	for i := range msg.Fields {
		if msg.Fields[i].Name == fieldName {
			return &msg.Fields[i]
		}
	}
	return nil
}
