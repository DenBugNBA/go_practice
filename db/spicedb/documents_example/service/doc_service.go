package service

import (
	"context"

	pb "go_practice/db/spicedb/documents_example/gen"
)

type DocumentService struct {
	pb.UnimplementedDocumentServiceServer
}

func (s *DocumentService) GetDocument(_ context.Context, req *pb.GetDocumentRequest) (*pb.GetDocumentResponse, error) {
	return &pb.GetDocumentResponse{
		Id:      req.Id,
		Content: "123",
	}, nil
}
