package main

import (
	"context"
	"grpc-gui/internal/grpcreflect"
	"grpc-gui/internal/utils"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	reflect, err := grpcreflect.NewReflector(ctx, "localhost:50051", &utils.GRPCConnectOptions{UseTLS: false})
	if err != nil {
		log.Fatalf("Failed to create reflector: %v", err)
	}
	defer reflect.Close()

	servicesInfo, err := reflect.GetAllServicesInfo()
	if err != nil {
		log.Fatalf("Failed to get services info: %v", err)
	}

	for _, s := range servicesInfo.Services {
		for _, m := range s.Methods {
			if m.Name != "ComplexCall" {
				continue
			}

			log.Println("=== Request Example ===")
			log.Println(string(m.RequestExample))

			log.Println("\n=== Request Schema (with enum values and oneof) ===")
			log.Println(string(m.RequestSchema))

			log.Println("\n=== Response Example ===")
			log.Println(string(m.ResponseExample))
		}
	}

}
