package authz

import (
	"context"
	"fmt"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	documents "go_practice/db/spicedb/documents_example/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorizer struct {
	client *authzed.Client
}

func NewAuthorizer(client *authzed.Client) *Authorizer {
	return &Authorizer{client: client}
}

func (a *Authorizer) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		userID, err := parseUserID(ctx)
		if err != nil {
			return nil, err
		}
		docReq, ok := req.(*documents.GetDocumentRequest)
		if !ok {
			return nil, fmt.Errorf("unknown request type for authz")
		}
		checkResp, err := a.client.CheckPermission(ctx, &v1.CheckPermissionRequest{
			Resource: &v1.ObjectReference{
				ObjectType: "document",
				ObjectId:   docReq.Id,
			},
			Permission: "read",
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: "user",
					ObjectId:   userID,
				},
			},
		})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "spicedb error: %v", err)
		}
		if checkResp.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
		return handler(ctx, req)
	}
}
