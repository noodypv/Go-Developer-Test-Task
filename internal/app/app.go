package app

import (
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"tages-go/internal/service"
	storagerpc "tages-go/internal/transport/grpc"
)

type App struct {
	grpcSrv *grpc.Server
	port    int
}

func New(port int, services service.Service, maxRead, maxUpDownLoad int) *App {
	gRPCServer := grpc.NewServer()

	storagerpc.Register(gRPCServer, services, maxRead, maxUpDownLoad)

	return &App{
		grpcSrv: gRPCServer,
		port:    port,
	}
}

func (a *App) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return err
	}

	log.Printf("storage service is running on port %d\n", a.port)

	if err := a.grpcSrv.Serve(l); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.grpcSrv.GracefulStop()
}
