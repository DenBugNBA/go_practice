package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	spicedbEndpoint = "localhost:50051"
	token           = "foobar"
)

func main() {
	client, err := authzed.NewClient(
		spicedbEndpoint,
		grpcutil.WithInsecureBearerToken(token),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}

	request := &pb.WriteRelationshipsRequest{
		Updates: []*pb.RelationshipUpdate{
			{
				Operation: pb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &pb.Relationship{
					Resource: &pb.ObjectReference{
						ObjectType: "document",
						ObjectId:   "1",
					},
					Relation: "reader",
					Subject: &pb.SubjectReference{
						Object: &pb.ObjectReference{
							ObjectType: "user",
							ObjectId:   "1",
						},
					},
				},
			},
		},
	}

	resp, err := client.WriteRelationships(context.Background(), request)
	if err != nil {
		log.Fatalf("failed to write relations: %s", err)
	}
	fmt.Println(resp.WrittenAt.Token)
}
