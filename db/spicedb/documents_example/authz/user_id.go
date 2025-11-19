package authz

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const userIDHeader = "x-user-id"

func parseUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	idVals := md.Get(userIDHeader)
	if len(idVals) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing user ID")
	}

	return idVals[0], nil
}
