package main

import (
	"context"
	"grpc-gui/internal/grpcreflect"
	"log"
	"time"

	"github.com/k0kubun/pp"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	reflect, err := grpcreflect.NewReflector(ctx, "localhost:50051", grpcreflect.ReflectorOptions{UseTLS: false})
	if err != nil {
		log.Fatalf("Failed to create reflector: %v", err)
	}
	defer reflect.Close()

	servicesInfo, err := reflect.GetAllServicesInfo()
	if err != nil {
		log.Fatalf("Failed to get services info: %v", err)
	}

	pp.Println(servicesInfo)
}
