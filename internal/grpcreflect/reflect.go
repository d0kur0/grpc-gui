package grpcreflect

import (
	"context"
	"encoding/json"
	"fmt"
	"grpc-gui/internal/utils"
	"strings"
	"time"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

type Reflector struct {
	conn   *grpc.ClientConn
	client *grpcreflect.Client
}

type EnumValueInfo struct {
	Name   string `json:"name"`
	Number int32  `json:"number"`
}

type FieldInfo struct {
 	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Number        int32           `json:"number"`
	Repeated      bool            `json:"repeated"`
	Optional      bool            `json:"optional"`
	Required      bool            `json:"required"`
	IsMap         bool            `json:"isMap"`
	IsEnum        bool            `json:"isEnum"`
	IsWellKnown   bool            `json:"isWellKnown"`
	WellKnownType string          `json:"wellKnownType,omitempty"`
	MapKey        string          `json:"mapKey"`
	MapValue      string          `json:"mapValue"`
	OneofGroup    string          `json:"oneofGroup,omitempty"`
	Message       *MessageInfo    `json:"message,omitempty"`
	EnumValues    []EnumValueInfo `json:"enumValues,omitempty"`
}

type MessageInfo struct {
	Name   string      `json:"name"`
	Fields []FieldInfo `json:"fields"`
}

type MethodInfo struct {
	Name                 string          `json:"name"`
	RequestType          string          `json:"requestType"`
	ResponseType         string          `json:"responseType"`
	Request              *MessageInfo    `json:"request,omitempty"`
	Response             *MessageInfo    `json:"response,omitempty"`
	RequestExample       json.RawMessage `json:"requestExample,omitempty"`
	RequestExampleString string          `json:"requestExampleString,omitempty"`
	ResponseExample      json.RawMessage `json:"responseExample,omitempty"`
	RequestSchema        json.RawMessage `json:"requestSchema,omitempty"`
}

type ServiceInfo struct {
	Name    string       `json:"name"`
	Methods []MethodInfo `json:"methods"`
}

type ServicesInfo struct {
	Services []ServiceInfo `json:"services"`
}

func NewReflector(ctx context.Context, url string, opts *utils.GRPCConnectOptions) (*Reflector, error) {
	conn, err := utils.CreateGRPCConnect(url, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connect: %w", err)
	}

	client := grpcreflect.NewClientAuto(ctx, conn)

	return &Reflector{
		conn:   conn,
		client: client,
	}, nil
}

func (r *Reflector) Close() error {
	r.client.Reset()
	return r.conn.Close()
}

func (r *Reflector) GetServiceDescriptor(serviceName string) (*desc.ServiceDescriptor, error) {
	return r.client.ResolveService(serviceName)
}

func IsSystemService(serviceName string) bool {
	return (strings.HasPrefix(serviceName, "grpc.reflection.") && strings.HasSuffix(serviceName, ".ServerReflection")) ||
		strings.HasPrefix(serviceName, "grpc.health.")
}

func (r *Reflector) getServiceMethodsLowLevel(ctx context.Context, serviceName string) ([]MethodInfo, error) {
	refClient := reflectpb.NewServerReflectionClient(r.conn)
	stream, err := refClient.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.CloseSend()

	err = stream.Send(&reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: serviceName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("failed to receive response: %w", err)
	}

	fdResp := resp.GetFileDescriptorResponse()
	if fdResp == nil {
		return nil, fmt.Errorf("no file descriptor response")
	}

	if len(fdResp.FileDescriptorProto) == 0 {
		return nil, fmt.Errorf("no file descriptors returned")
	}

	allFiles := make(map[string]*descriptorpb.FileDescriptorProto)
	for _, fdBytes := range fdResp.FileDescriptorProto {
		fdProto := &descriptorpb.FileDescriptorProto{}
		err = proto.Unmarshal(fdBytes, fdProto)
		if err != nil {
			continue
		}
		allFiles[fdProto.GetName()] = fdProto
	}

	mainFd := &descriptorpb.FileDescriptorProto{}
	err = proto.Unmarshal(fdResp.FileDescriptorProto[0], mainFd)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal descriptor: %w", err)
	}

	var methods []MethodInfo
	for _, svc := range mainFd.GetService() {
		for _, method := range svc.GetMethod() {
			methodInfo := MethodInfo{
				Name:         method.GetName(),
				RequestType:  method.GetInputType(),
				ResponseType: method.GetOutputType(),
			}

			requestMsg := r.findAndBuildMessage(method.GetInputType(), allFiles)
			responseMsg := r.findAndBuildMessage(method.GetOutputType(), allFiles)

			if requestMsg != nil {
				methodInfo.Request = requestMsg
				requestExample, _ := GenerateJSONExample(requestMsg)
				methodInfo.RequestExample = json.RawMessage(requestExample)
				methodInfo.RequestExampleString = GenerateJSONExampleWithComments(requestMsg)
				requestSchema, _ := GenerateRequestSchema(requestMsg)
				methodInfo.RequestSchema = json.RawMessage(requestSchema)
			}

			if responseMsg != nil {
				methodInfo.Response = responseMsg
				responseExample, _ := GenerateJSONExample(responseMsg)
				methodInfo.ResponseExample = json.RawMessage(responseExample)
			}

			methods = append(methods, methodInfo)
		}
	}

	return methods, nil
}

func (r *Reflector) findAndBuildMessage(typeName string, files map[string]*descriptorpb.FileDescriptorProto) *MessageInfo {
	cleanName := strings.TrimPrefix(typeName, ".")

	for _, fd := range files {
		pkg := fd.GetPackage()
		for _, msg := range fd.GetMessageType() {
			fullName := pkg + "." + msg.GetName()
			if fullName == cleanName {
				return r.buildMessageFromProto(msg, fd, files)
			}
		}
	}

	return nil
}

func (r *Reflector) findEnumValues(typeName string, files map[string]*descriptorpb.FileDescriptorProto) []EnumValueInfo {
	cleanName := strings.TrimPrefix(typeName, ".")

	for _, fd := range files {
		pkg := fd.GetPackage()
		for _, enum := range fd.GetEnumType() {
			fullName := pkg + "." + enum.GetName()
			if fullName == cleanName {
				values := make([]EnumValueInfo, 0, len(enum.GetValue()))
				for _, val := range enum.GetValue() {
					values = append(values, EnumValueInfo{
						Name:   val.GetName(),
						Number: val.GetNumber(),
					})
				}
				return values
			}
		}
	}

	return nil
}

func (r *Reflector) buildMessageFromProto(msgProto *descriptorpb.DescriptorProto, fd *descriptorpb.FileDescriptorProto, files map[string]*descriptorpb.FileDescriptorProto) *MessageInfo {
	pkg := fd.GetPackage()
	fullName := pkg + "." + msgProto.GetName()

	info := &MessageInfo{
		Name:   fullName,
		Fields: []FieldInfo{},
	}

	for _, field := range msgProto.GetField() {
		fieldInfo := FieldInfo{
			Name:     field.GetName(),
			Number:   field.GetNumber(),
			Repeated: field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED,
			Optional: field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL,
			Required: field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REQUIRED,
		}

		if field.OneofIndex != nil {
			oneofDecl := msgProto.GetOneofDecl()[*field.OneofIndex]
			fieldInfo.OneofGroup = oneofDecl.GetName()
		}

		fieldInfo.Type = r.getFieldType(field)

		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE && field.GetTypeName() != "" {
			typeName := strings.TrimPrefix(field.GetTypeName(), ".")
			isWellKnown, wellKnownType := isWellKnownType(typeName)
			
			if isWellKnown {
				fieldInfo.IsWellKnown = true
				fieldInfo.WellKnownType = wellKnownType
			} else {
				nestedMsg := r.findAndBuildMessage(field.GetTypeName(), files)
				if nestedMsg != nil {
					fieldInfo.Message = nestedMsg
				}
			}
		}

		if field.GetType() == descriptorpb.FieldDescriptorProto_TYPE_ENUM && field.GetTypeName() != "" {
			fieldInfo.IsEnum = true
			enumValues := r.findEnumValues(field.GetTypeName(), files)
			if len(enumValues) > 0 {
				fieldInfo.EnumValues = enumValues
			}
		}

		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
}

func (r *Reflector) getFieldType(field *descriptorpb.FieldDescriptorProto) string {
	switch field.GetType() {
	case descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		return "double"
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT:
		return "float"
	case descriptorpb.FieldDescriptorProto_TYPE_INT64:
		return "int64"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		return "uint64"
	case descriptorpb.FieldDescriptorProto_TYPE_INT32:
		return "int32"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		return "fixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		return "fixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		return "string"
	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		return "bytes"
	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		return "uint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		return "sfixed32"
	case descriptorpb.FieldDescriptorProto_TYPE_SFIXED64:
		return "sfixed64"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT32:
		return "sint32"
	case descriptorpb.FieldDescriptorProto_TYPE_SINT64:
		return "sint64"
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if field.GetTypeName() != "" {
			return strings.TrimPrefix(field.GetTypeName(), ".")
		}
		return "message"
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		if field.GetTypeName() != "" {
			return strings.TrimPrefix(field.GetTypeName(), ".")
		}
		return "enum"
	default:
		return "unknown"
	}
}

func (r *Reflector) GetAllServicesInfo() (*ServicesInfo, error) {
	serviceNames, err := r.client.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []ServiceInfo

	for _, serviceName := range serviceNames {
		serviceInfo := ServiceInfo{
			Name:    serviceName,
			Methods: []MethodInfo{},
		}

		serviceDesc, err := r.client.ResolveService(serviceName)
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			lowLevelMethods, lowLevelErr := r.getServiceMethodsLowLevel(ctx, serviceName)
			cancel()

			if lowLevelErr == nil {
				serviceInfo.Methods = lowLevelMethods
			}

			services = append(services, serviceInfo)
			continue
		}

		unwrapped := serviceDesc.UnwrapService()
		methods := unwrapped.Methods()
		for i := 0; i < methods.Len(); i++ {
			method := methods.Get(i)

			requestMsg := extractMessageInfo(method.Input())
			responseMsg := extractMessageInfo(method.Output())

			requestExample, _ := GenerateJSONExample(requestMsg)
			responseExample, _ := GenerateJSONExample(responseMsg)
			requestSchema, _ := GenerateRequestSchema(requestMsg)
			requestExampleString := GenerateJSONExampleWithComments(requestMsg)

			serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
				Name:                 string(method.Name()),
				RequestType:          string(method.Input().FullName()),
				ResponseType:         string(method.Output().FullName()),
				Request:              requestMsg,
				Response:             responseMsg,
				RequestExample:       json.RawMessage(requestExample),
				RequestExampleString: requestExampleString,
				ResponseExample:      json.RawMessage(responseExample),
				RequestSchema:        json.RawMessage(requestSchema),
			})
		}

		services = append(services, serviceInfo)
	}

	return &ServicesInfo{Services: services}, nil
}

func extractMessageInfo(msgDesc protoreflect.MessageDescriptor) *MessageInfo {
	return extractMessageInfoRecursive(msgDesc, make(map[string]bool))
}

func extractMessageInfoRecursive(msgDesc protoreflect.MessageDescriptor, visited map[string]bool) *MessageInfo {
	if msgDesc == nil {
		return nil
	}

	fullName := string(msgDesc.FullName())
	if visited[fullName] {
		return &MessageInfo{
			Name:   fullName,
			Fields: []FieldInfo{},
		}
	}
	visited[fullName] = true

	info := &MessageInfo{
		Name:   fullName,
		Fields: []FieldInfo{},
	}

	fields := msgDesc.Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		fieldType := fieldKindToString(field.Kind())
		var nestedMsg *MessageInfo
		isMap := false
		mapKey := ""
		mapValue := ""

		isWellKnown := false
		wellKnownType := ""

		if field.IsMap() {
			isMap = true
			mapKey = fieldKindToString(field.MapKey().Kind())
			if field.MapValue().Message() != nil {
				mapValue = string(field.MapValue().Message().FullName())
				nestedMsg = extractMessageInfoRecursive(field.MapValue().Message(), visited)
			} else if field.MapValue().Enum() != nil {
				mapValue = string(field.MapValue().Enum().FullName())
			} else {
				mapValue = fieldKindToString(field.MapValue().Kind())
			}
			fieldType = fmt.Sprintf("map<%s, %s>", mapKey, mapValue)
		} else if field.Message() != nil {
			fieldType = string(field.Message().FullName())
			isWellKnown, wellKnownType = isWellKnownType(fieldType)
			if !isWellKnown {
				nestedMsg = extractMessageInfoRecursive(field.Message(), visited)
			}
		} else if field.Enum() != nil {
			fieldType = string(field.Enum().FullName())
		}

		var enumValues []EnumValueInfo
		isEnum := false
		if field.Enum() != nil {
			isEnum = true
			enumDesc := field.Enum()
			values := enumDesc.Values()
			for i := 0; i < values.Len(); i++ {
				value := values.Get(i)
				enumValues = append(enumValues, EnumValueInfo{
					Name:   string(value.Name()),
					Number: int32(value.Number()),
				})
			}
		} else if field.IsMap() && field.MapValue().Enum() != nil {
			enumDesc := field.MapValue().Enum()
			values := enumDesc.Values()
			for i := 0; i < values.Len(); i++ {
				value := values.Get(i)
				enumValues = append(enumValues, EnumValueInfo{
					Name:   string(value.Name()),
					Number: int32(value.Number()),
				})
			}
		}

		oneofGroup := ""
		if oneof := field.ContainingOneof(); oneof != nil {
			oneofGroup = string(oneof.Name())
		}

		fieldInfo := FieldInfo{
			Name:          string(field.Name()),
			Type:          fieldType,
			Number:        int32(field.Number()),
			Repeated:      field.Cardinality() == protoreflect.Repeated && !field.IsMap(),
			Optional:      field.Cardinality() == protoreflect.Optional,
			Required:      field.Cardinality() == protoreflect.Required,
			IsMap:         isMap,
			IsEnum:        isEnum,
			IsWellKnown:   isWellKnown,
			WellKnownType: wellKnownType,
			MapKey:        mapKey,
			MapValue:      mapValue,
			OneofGroup:    oneofGroup,
			Message:       nestedMsg,
			EnumValues:    enumValues,
		}
		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
}

func isWellKnownType(typeName string) (bool, string) {
	switch typeName {
	case "google.protobuf.Timestamp":
		return true, "timestamp"
	case "google.protobuf.Duration":
		return true, "duration"
	case "google.protobuf.Any":
		return true, "any"
	case "google.protobuf.Struct":
		return true, "struct"
	case "google.protobuf.Value":
		return true, "value"
	case "google.protobuf.ListValue":
		return true, "list_value"
	case "google.protobuf.Empty":
		return true, "empty"
	default:
		return false, ""
	}
}

func fieldKindToString(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind:
		return "int32"
	case protoreflect.Sint32Kind:
		return "sint32"
	case protoreflect.Sfixed32Kind:
		return "sfixed32"
	case protoreflect.Int64Kind:
		return "int64"
	case protoreflect.Sint64Kind:
		return "sint64"
	case protoreflect.Sfixed64Kind:
		return "sfixed64"
	case protoreflect.Uint32Kind:
		return "uint32"
	case protoreflect.Fixed32Kind:
		return "fixed32"
	case protoreflect.Uint64Kind:
		return "uint64"
	case protoreflect.Fixed64Kind:
		return "fixed64"
	case protoreflect.FloatKind:
		return "float"
	case protoreflect.DoubleKind:
		return "double"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "bytes"
	case protoreflect.MessageKind:
		return "message"
	case protoreflect.EnumKind:
		return "enum"
	case protoreflect.GroupKind:
		return "group"
	default:
		return kind.String()
	}
}

func GenerateJSONExample(msg *MessageInfo) ([]byte, error) {
	if msg == nil {
		return []byte("{}"), nil
	}

	data := GenerateJSONValue(msg, make(map[string]bool))
	return json.MarshalIndent(data, "", "  ")
}

func GenerateJSONExampleWithComments(msg *MessageInfo) string {
	if msg == nil {
		return "{}"
	}
	return generateJSONWithComments(msg, 0, make(map[string]bool), make(map[string]bool))
}

func generateJSONWithComments(msg *MessageInfo, indent int, visited map[string]bool, processedOneofs map[string]bool) string {
	if msg == nil {
		return "{}"
	}

	fullName := msg.Name
	if visited[fullName] {
		return "{}"
	}
	visited[fullName] = true
	defer delete(visited, fullName)

	oneofGroups := make(map[string]int)
	for _, field := range msg.Fields {
		if field.OneofGroup != "" {
			oneofGroups[field.OneofGroup]++
		}
	}

	var result strings.Builder
	result.WriteString("{\n")

	indentStr := strings.Repeat("  ", indent+1)
	first := true

	for _, field := range msg.Fields {
		if !first {
			result.WriteString(",\n")
		}
		first = false

		if field.OneofGroup != "" {
			oneofKey := fullName + "." + field.OneofGroup
			if !processedOneofs[oneofKey] {
				processedOneofs[oneofKey] = true

				if oneofGroups[field.OneofGroup] > 1 {
					result.WriteString(indentStr)
					result.WriteString(fmt.Sprintf("// oneof %s (choose one):\n", field.OneofGroup))
				}
			}
		}

		result.WriteString(indentStr)
		result.WriteString(fmt.Sprintf(`"%s": `, field.Name))
		result.WriteString(generateFieldValueWithComments(field, indent+1, visited, processedOneofs))
	}

	result.WriteString("\n")
	result.WriteString(strings.Repeat("  ", indent))
	result.WriteString("}")

	return result.String()
}

func generateFieldValueWithComments(field FieldInfo, indent int, visited map[string]bool, processedOneofs map[string]bool) string {
	if field.Repeated {
		return "[]"
	}

	if field.IsMap {
		return "{}"
	}

	if field.IsWellKnown {
		switch field.WellKnownType {
		case "timestamp":
			return `"2026-02-05T14:05:47Z"`
		case "duration":
			return `"1.5s"`
		case "empty":
			return "{}"
		case "struct":
			return "{}"
		case "value":
			return "null"
		case "list_value":
			return "[]"
		case "any":
			return `{"@type": ""}`
		}
	}

	if field.Message != nil {
		return generateJSONWithComments(field.Message, indent, visited, processedOneofs)
	}

	if field.IsEnum && len(field.EnumValues) > 0 {
		return fmt.Sprintf(`"%s"`, field.EnumValues[0].Name)
	}

	switch field.Type {
	case "string":
		return `""`
	case "int32", "int64", "uint32", "uint64", "sint32", "sint64",
		"fixed32", "fixed64", "sfixed32", "sfixed64":
		return "0"
	case "double", "float":
		return "0.0"
	case "bool":
		return "false"
	case "bytes":
		return `""`
	default:
		return `""`
	}
}

func GenerateJSONValue(msg *MessageInfo, visited map[string]bool) map[string]interface{} {
	if msg == nil {
		return nil
	}

	fullName := msg.Name
	if visited[fullName] {
		return map[string]interface{}{}
	}
	visited[fullName] = true

	result := make(map[string]interface{})

	for _, field := range msg.Fields {
		result[field.Name] = generateFieldValue(field, visited)
	}

	delete(visited, fullName)
	return result
}

func generateFieldValue(field FieldInfo, visited map[string]bool) interface{} {
	if field.IsMap {
		return map[string]interface{}{}
	}

	if field.Repeated {
		if field.Message != nil {
			return []interface{}{GenerateJSONValue(field.Message, visited)}
		}
		return []interface{}{getDefaultValueForField(field)}
	}

	if field.Message != nil {
		return GenerateJSONValue(field.Message, visited)
	}

	return getDefaultValueForField(field)
}

func getDefaultValueForType(typeStr string) interface{} {
	switch typeStr {
	case "bool":
		return false
	case "int32", "sint32", "sfixed32":
		return 0
	case "int64", "sint64", "sfixed64":
		return int64(0)
	case "uint32", "fixed32":
		return uint32(0)
	case "uint64", "fixed64":
		return uint64(0)
	case "float":
		return float32(0.0)
	case "double":
		return 0.0
	case "string":
		return ""
	case "bytes":
		return ""
	default:
		if len(typeStr) > 0 {
			return "UNKNOWN"
		}
		return ""
	}
}

func getDefaultValueForField(field FieldInfo) interface{} {
	if field.IsWellKnown {
		switch field.WellKnownType {
		case "timestamp":
			return "2026-02-05T14:05:47Z"
		case "duration":
			return "1.5s"
		case "empty":
			return map[string]interface{}{}
		case "struct":
			return map[string]interface{}{}
		case "value":
			return nil
		case "list_value":
			return []interface{}{}
		case "any":
			return map[string]interface{}{"@type": ""}
		}
	}
	
	if len(field.EnumValues) > 0 {
		return field.EnumValues[0].Name
	}
	return getDefaultValueForType(field.Type)
}

func GenerateRequestSchema(msg *MessageInfo) ([]byte, error) {
	if msg == nil {
		return []byte("{}"), nil
	}

	schema := GenerateSchemaValue(msg, make(map[string]bool))
	return json.MarshalIndent(schema, "", "  ")
}

func GenerateSchemaValue(msg *MessageInfo, visited map[string]bool) map[string]interface{} {
	if msg == nil {
		return nil
	}

	fullName := msg.Name
	if visited[fullName] {
		return map[string]interface{}{}
	}
	visited[fullName] = true

	result := make(map[string]interface{})

	for _, field := range msg.Fields {
		fieldSchema := make(map[string]interface{})
		fieldSchema["type"] = field.Type
		fieldSchema["value"] = generateFieldValue(field, visited)

		if field.Repeated {
			fieldSchema["repeated"] = true
		}
		if field.Optional {
			fieldSchema["optional"] = true
		}
		if field.Required {
			fieldSchema["required"] = true
		}
		if field.IsMap {
			fieldSchema["isMap"] = true
			fieldSchema["mapKey"] = field.MapKey
			fieldSchema["mapValue"] = field.MapValue
		}
		if len(field.EnumValues) > 0 {
			fieldSchema["isEnum"] = true
			enumValues := make([]string, 0, len(field.EnumValues))
			for _, ev := range field.EnumValues {
				enumValues = append(enumValues, ev.Name)
			}
			fieldSchema["enumValues"] = enumValues
		}
		if field.OneofGroup != "" {
			fieldSchema["oneofGroup"] = field.OneofGroup
		}
		if field.Message != nil {
			fieldSchema["message"] = GenerateSchemaValue(field.Message, visited)
		}

		result[field.Name] = fieldSchema
	}

	delete(visited, fullName)
	return result
}
