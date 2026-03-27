package main

import (
	"go_practice/go_course/7_async_logger/service"

	"google.golang.org/grpc"
)

func newAdminService() *adminService {
	return &adminService{}
}

type adminService struct {
	service.UnimplementedAdminServer
}

func (a adminService) Logging(nothing *service.Nothing, str grpc.ServerStreamingServer[service.Event]) error {
	//TODO implement me
	panic("implement me")
}

func (a adminService) Statistics(interval *service.StatInterval, str grpc.ServerStreamingServer[service.Stat]) error {
	//TODO implement me
	panic("implement me")
}
