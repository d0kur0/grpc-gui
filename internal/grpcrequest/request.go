package grpcrequest

import (
	"context"
	"fmt"
	"time"

	"grpc-gui/internal/grpcreflect"

	"github.com/jhump/protoreflect/dynamic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func DoGRPCRequest(address, service, method, payload string, requestHeaders, contextValues map[string]string) (string, codes.Code, map[string][]string, error) {
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

	reflector, err := grpcreflect.NewReflector(ctx, address, &grpcreflect.ReflectorOptions{UseTLS: false})
	if err != nil {
		return "", codes.Unknown, nil, fmt.Errorf("failed to create reflector: %w", err)
	}
	defer reflector.Close()

	serviceDesc, err := reflector.GetServiceDescriptor(service)
	if err != nil {
		return "", codes.NotFound, nil, fmt.Errorf("failed to resolve service: %w", err)
	}

	methodDesc := serviceDesc.FindMethodByName(method)
	if methodDesc == nil {
		return "", codes.NotFound, nil, fmt.Errorf("method %s not found", method)
	}

	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", codes.Unavailable, nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer conn.Close()

	reqMsg := dynamic.NewMessage(methodDesc.GetInputType())

	if payload != "" {
		if err := reqMsg.UnmarshalJSON([]byte(payload)); err != nil {
			return "", codes.InvalidArgument, nil, fmt.Errorf("failed to parse payload: %w", err)
		}
	}

	methodPath := fmt.Sprintf("/%s/%s", service, method)
	respMsg := dynamic.NewMessage(methodDesc.GetOutputType())

	var responseHeaders metadata.MD
	var trailer metadata.MD
	err = conn.Invoke(ctx, methodPath, reqMsg, respMsg, grpc.Header(&responseHeaders), grpc.Trailer(&trailer))
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if len(trailer) > 0 {
				return "", st.Code(), trailer, err
			}
			return "", st.Code(), nil, err
		}
		return "", codes.Unknown, nil, fmt.Errorf("rpc call failed: %w", err)
	}

	respJSON, err := respMsg.MarshalJSON()
	if err != nil {
		return "", codes.Internal, nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(respJSON), codes.OK, responseHeaders, nil
}
