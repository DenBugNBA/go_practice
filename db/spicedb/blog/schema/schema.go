package main

import (
	"context"
	"log"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const schema = `definition test/user {}
definition test/post {
	relation reader: test/user
	relation writer: test/user
	permission read = reader + writer
	permission write = writer
}`

const spicedbEndpoint = "localhost:50051"

func main() {
	client, err := authzed.NewClient(
		spicedbEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithBearerToken("sometoken"),
	)
	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}

	request := &pb.WriteSchemaRequest{Schema: schema}
	_, err = client.WriteSchema(context.Background(), request)
	if err != nil {
		log.Fatalf("failed to write schema: %s", err)
	}
}
