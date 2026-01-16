package main

import (
	"log"
	"net"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"go_practice/db/spicedb/documents_example/authz"
	pb "go_practice/db/spicedb/documents_example/gen"
	"go_practice/db/spicedb/documents_example/service"
)

const (
	spicedbHost = "localhost:50051"
	token       = "foobar"
)

func main() {
	c, err := authzed.NewClient(
		spicedbHost,
		grpcutil.WithInsecureBearerToken(token),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	authorizer := authz.NewAuthorizer(c)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authorizer.Unary()),
	)
	reflection.Register(server)

	pb.RegisterDocumentServiceServer(server, &service.DocumentService{})

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	if err := server.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
