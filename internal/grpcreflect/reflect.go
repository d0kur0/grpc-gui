package grpcreflect

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Reflector struct {
	conn   *grpc.ClientConn
	client *grpcreflect.Client
}

type ReflectorOptions struct {
	UseTLS   bool
	Insecure bool
}

type FieldInfo struct {
	Name     string
	Type     string
	Number   int32
	Repeated bool
	Optional bool
	Required bool
	IsMap    bool
	MapKey   string
	MapValue string
	Message  *MessageInfo
}

type MessageInfo struct {
	Name   string
	Fields []FieldInfo
}

type MethodInfo struct {
	Name         string
	RequestType  string
	ResponseType string
	Request      *MessageInfo
	Response     *MessageInfo
}

type ServiceInfo struct {
	Name    string
	Methods []MethodInfo
}

type ServicesInfo struct {
	Services []ServiceInfo
}

func NewReflector(ctx context.Context, url string, opts ReflectorOptions) (*Reflector, error) {
	var creds credentials.TransportCredentials

	if opts.UseTLS {
		if opts.Insecure {
			creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		} else {
			creds = credentials.NewTLS(&tls.Config{})
		}
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
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

		info.Fields = append(info.Fields, FieldInfo{
			Name:     string(field.Name()),
			Type:     fieldType,
			Number:   int32(field.Number()),
			Repeated: field.Cardinality() == protoreflect.Repeated && !field.IsMap(),
			Optional: field.Cardinality() == protoreflect.Optional,
			Required: field.Cardinality() == protoreflect.Required,
			IsMap:    isMap,
			MapKey:   mapKey,
			MapValue: mapValue,
			Message:  nestedMsg,
		})
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
