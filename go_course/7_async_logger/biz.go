package main

import (
	"context"

	"go_practice/go_course/7_async_logger/service"
)

func newBizService() *bizService {
	return &bizService{}
}

type bizService struct {
	service.UnimplementedBizServer
}

func (b bizService) Check(_ context.Context, _ *service.Nothing) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}

func (b bizService) Add(_ context.Context, _ *service.Nothing) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}

func (b bizService) Test(_ context.Context, _ *service.Nothing) (*service.Nothing, error) {
	return &service.Nothing{}, nil
}
