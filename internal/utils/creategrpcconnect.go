package utils

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCConnectOptions struct {
	UseTLS   bool
	Insecure bool
}

func CreateGRPCConnect(address string, opts *GRPCConnectOptions) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials

	if opts != nil && opts.UseTLS {
		if opts.Insecure {
			creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		} else {
			creds = credentials.NewTLS(&tls.Config{})
		}
	} else {
		creds = insecure.NewCredentials()
	}
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	return conn, nil
}
