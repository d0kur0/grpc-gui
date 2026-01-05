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

	reflect, err := grpcreflect.NewReflector(ctx, "localhost:50051", &grpcreflect.ReflectorOptions{UseTLS: false})
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

			if m.Name != "GetUser" {
				continue
			}

			j := grpcreflect.GenerateJSONValue(m.Response, make(map[string]bool))
			pp.Println(j)

			a, _ := grpcreflect.GenerateJSONExample(m.Response)

			log.Println(string(a))
		}
	}

}
