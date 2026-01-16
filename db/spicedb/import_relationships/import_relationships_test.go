package import_relationships

import (
	"io"
	"testing"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	endpoint = "localhost:50051"
	token    = "key"
	schema   = `
definition user {}

definition document {
    relation owner: user
}
`
)

func TestImportBulkRelationships(t *testing.T) {
	c, err := authzed.NewClient(
		endpoint,
		grpcutil.WithInsecureBearerToken(token),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("create spiceDB client: %v", err)
	}

	_, err = c.WriteSchema(t.Context(), &pb.WriteSchemaRequest{Schema: schema})
	if err != nil {
		t.Fatalf("write schema: %v", err)
	}

	t.Run("bulk import without sends", func(t *testing.T) {
		s, err := c.ImportBulkRelationships(t.Context())
		if err != nil {
			t.Fatalf("create import stream: %v", err)
		}
		_, err = s.CloseAndRecv()
		if err != nil {
			t.Fatalf("close and resv: %v", err)
		}
	})

	t.Run("bulk import with conflict", func(t *testing.T) {
		_, err = c.WriteRelationships(t.Context(), &pb.WriteRelationshipsRequest{
			Updates: []*pb.RelationshipUpdate{
				{
					Operation: pb.RelationshipUpdate_OPERATION_CREATE,
					Relationship: &pb.Relationship{
						Resource: &pb.ObjectReference{
							ObjectType: "document",
							ObjectId:   "1",
						},
						Relation: "owner",
						Subject: &pb.SubjectReference{
							Object: &pb.ObjectReference{
								ObjectType: "user",
								ObjectId:   "1",
							},
						}},
				},
			},
		})
		if err != nil {
			t.Fatalf("write relationships: %v", err)
		}

		s, err := c.ImportBulkRelationships(t.Context())
		if err != nil {
			t.Fatalf("create import stream: %v", err)
		}

		req := &pb.ImportBulkRelationshipsRequest{
			Relationships: []*pb.Relationship{
				{
					Resource: &pb.ObjectReference{
						ObjectType: "document",
						ObjectId:   "1",
					},
					Relation: "owner",
					Subject: &pb.SubjectReference{
						Object: &pb.ObjectReference{
							ObjectType: "user",
							ObjectId:   "1",
						},
					},
				},
			},
		}
		if err = s.Send(req); err != nil {
			t.Fatalf("send import req: %v", err)
		}

		req = &pb.ImportBulkRelationshipsRequest{
			Relationships: []*pb.Relationship{
				{
					Resource: &pb.ObjectReference{
						ObjectType: "document",
						ObjectId:   "2",
					},
					Relation: "owner",
					Subject: &pb.SubjectReference{
						Object: &pb.ObjectReference{
							ObjectType: "user",
							ObjectId:   "1",
						},
					},
				},
			},
		}
		if err = s.Send(req); err != nil {
			t.Fatalf("send import req: %v", err)
		}

		_, err = s.CloseAndRecv()
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("status from error: %v", st)
		}
		if st.Code() != codes.AlreadyExists {
			t.Fatalf("got status code %v, wanted %v", st.Code(), codes.AlreadyExists)
		}

		readClient, err := c.ReadRelationships(t.Context(), &pb.ReadRelationshipsRequest{
			Consistency: &pb.Consistency{
				Requirement: &pb.Consistency_FullyConsistent{
					FullyConsistent: true,
				},
			},
			RelationshipFilter: &pb.RelationshipFilter{
				ResourceType: "document",
			},
		})
		if err != nil {
			t.Fatalf("init read relationships client: %v", err)
		}

		relationships := make([]*pb.Relationship, 0)
		for {
			resp, err := readClient.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("read relationships: %v", err)
			}
			relationships = append(relationships, resp.GetRelationship())
		}

		require.Len(t, relationships, 1)
	})
}
