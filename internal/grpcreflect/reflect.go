package grpcreflect

import (
	"context"
	"encoding/json"
	"fmt"
	"grpc-gui/internal/utils"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
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
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Number     int32           `json:"number"`
	Repeated   bool            `json:"repeated"`
	Optional   bool            `json:"optional"`
	Required   bool            `json:"required"`
	IsMap      bool            `json:"isMap"`
	IsEnum     bool            `json:"isEnum"`
	MapKey     string          `json:"mapKey"`
	MapValue   string          `json:"mapValue"`
	Message    *MessageInfo    `json:"message,omitempty"`
	EnumValues []EnumValueInfo `json:"enumValues,omitempty"`
}

type MessageInfo struct {
	Name   string      `json:"name"`
	Fields []FieldInfo `json:"fields"`
}

type MethodInfo struct {
	Name         string       `json:"name"`
	RequestType  string       `json:"requestType"`
	ResponseType string       `json:"responseType"`
	Request      *MessageInfo `json:"request,omitempty"`
	Response     *MessageInfo `json:"response,omitempty"`
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

func (r *Reflector) GetAllServicesInfo() (*ServicesInfo, error) {
	serviceNames, err := r.client.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []ServiceInfo

	for _, serviceName := range serviceNames {
		serviceDesc, err := r.client.ResolveService(serviceName)
		if err != nil {
			continue
		}

		serviceInfo := ServiceInfo{
			Name:    serviceName,
			Methods: []MethodInfo{},
		}

		unwrapped := serviceDesc.UnwrapService()
		methods := unwrapped.Methods()
		for i := 0; i < methods.Len(); i++ {
			method := methods.Get(i)

			requestMsg := extractMessageInfo(method.Input())
			responseMsg := extractMessageInfo(method.Output())

			serviceInfo.Methods = append(serviceInfo.Methods, MethodInfo{
				Name:         string(method.Name()),
				RequestType:  string(method.Input().FullName()),
				ResponseType: string(method.Output().FullName()),
				Request:      requestMsg,
				Response:     responseMsg,
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
			nestedMsg = extractMessageInfoRecursive(field.Message(), visited)
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

		fieldInfo := FieldInfo{
			Name:       string(field.Name()),
			Type:       fieldType,
			Number:     int32(field.Number()),
			Repeated:   field.Cardinality() == protoreflect.Repeated && !field.IsMap(),
			Optional:   field.Cardinality() == protoreflect.Optional,
			Required:   field.Cardinality() == protoreflect.Required,
			IsMap:      isMap,
			IsEnum:     isEnum,
			MapKey:     mapKey,
			MapValue:   mapValue,
			Message:    nestedMsg,
			EnumValues: enumValues,
		}
		info.Fields = append(info.Fields, fieldInfo)
	}

	return info
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
	if len(field.EnumValues) > 0 {
		return field.EnumValues[0].Name
	}
	return getDefaultValueForType(field.Type)
}
