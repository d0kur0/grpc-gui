package grpcrequest

import (
	"context"
	"fmt"
	"time"

	"grpc-gui/internal/grpcreflect"
	"grpc-gui/internal/utils"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

func getMethodDescriptorLowLevel(ctx context.Context, conn *grpc.ClientConn, serviceName, methodName string) (*desc.MethodDescriptor, error) {
	refClient := reflectpb.NewServerReflectionClient(conn)
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
	if fdResp == nil || len(fdResp.FileDescriptorProto) == 0 {
		return nil, fmt.Errorf("no file descriptors returned")
	}

	allFds := make(map[string]*descriptorpb.FileDescriptorProto)

	stubValidate := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("protoc-gen-validate/validate/validate.proto"),
		Package: proto.String("validate"),
		Syntax:  proto.String("proto3"),
	}
	allFds["protoc-gen-validate/validate/validate.proto"] = stubValidate

	for _, fdBytes := range fdResp.FileDescriptorProto {
		fdProto := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(fdBytes, fdProto); err != nil {
			continue
		}
		allFds[fdProto.GetName()] = fdProto
	}

	if len(allFds) == 0 {
		return nil, fmt.Errorf("failed to parse file descriptors")
	}

	var files protoregistry.Files

	maxAttempts := len(allFds)
	created := make(map[string]bool)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		progressMade := false

		for name, fdProto := range allFds {
			if created[name] {
				continue
			}

			fd, err := protodesc.NewFile(fdProto, &files)
			if err != nil {
				continue
			}

			files.RegisterFile(fd)
			created[name] = true
			progressMade = true
		}

		if !progressMade {
			break
		}
	}

	var methodDesc *desc.MethodDescriptor
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		wrappedFd, err := desc.WrapFile(fd)
		if err != nil {
			return true
		}

		for _, svc := range wrappedFd.GetServices() {
			if svc.GetFullyQualifiedName() == serviceName {
				for _, method := range svc.GetMethods() {
					if method.GetName() == methodName {
						methodDesc = method
						return false
					}
				}
			}
		}
		return true
	})

	if methodDesc != nil {
		return methodDesc, nil
	}

	return nil, fmt.Errorf("method %s not found in service %s", methodName, serviceName)
}

func DoGRPCRequest(address, service, method, payload string, requestHeaders, contextValues map[string]string, opts *utils.GRPCConnectOptions) (string, codes.Code, map[string][]string, int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if len(contextValues) > 0 {
		for k, v := range contextValues {
			ctx = context.WithValue(ctx, k, v)
		}
	}

	if len(requestHeaders) > 0 {
		md := metadata.New(requestHeaders)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	conn, err := utils.CreateGRPCConnect(address, opts)
	if err != nil {
		return "", codes.Unavailable, nil, 0, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	reflector, err := grpcreflect.NewReflector(ctx, address, opts)
	if err != nil {
		return "", codes.Unknown, nil, 0, fmt.Errorf("failed to create reflector: %w", err)
	}
	defer reflector.Close()

	var methodDesc *desc.MethodDescriptor

	serviceDesc, err := reflector.GetServiceDescriptor(service)
	if err != nil {
		methodDesc, err = getMethodDescriptorLowLevel(ctx, conn, service, method)
		if err != nil {
			return "", codes.NotFound, nil, 0, fmt.Errorf("failed to resolve method: %w", err)
		}
	} else {
		methodDesc = serviceDesc.FindMethodByName(method)
		if methodDesc == nil {
			return "", codes.NotFound, nil, 0, fmt.Errorf("method %s not found", method)
		}
	}

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())

	if payload != "" {
		if err := reqMsg.UnmarshalJSON([]byte(payload)); err != nil {
			return "", codes.InvalidArgument, nil, 0, fmt.Errorf("failed to parse payload: %w", err)
		}
	}

	methodPath := fmt.Sprintf("/%s/%s", service, method)
	respMsg := dynamic.NewMessage(methodDesc.GetOutputType())

	var responseHeaders metadata.MD
	var trailer metadata.MD

	startTime := time.Now()
	err = conn.Invoke(ctx, methodPath, reqMsg, respMsg, grpc.Header(&responseHeaders), grpc.Trailer(&trailer))
	executionTime := int32(time.Since(startTime).Milliseconds())

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if len(trailer) > 0 {
				return "", st.Code(), trailer, executionTime, err
			}
			return "", st.Code(), nil, executionTime, err
		}
		return "", codes.Unknown, nil, executionTime, fmt.Errorf("rpc call failed: %w", err)
	}

	respJSON, err := respMsg.MarshalJSON()
	if err != nil {
		return "", codes.Internal, nil, executionTime, fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(respJSON), codes.OK, responseHeaders, executionTime, nil
}
