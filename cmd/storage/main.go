package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tages-go/internal/app"
	"tages-go/internal/service"
)

const (
	port                    = 8900
	maxConcurrentRead       = 100
	maxConcurrentUpDownLoad = 10
	dataPath                = ".\\tmp"
)

func main() {
	services := service.NewServices(dataPath)

	a := app.New(port, *services, maxConcurrentRead, maxConcurrentUpDownLoad)

	go func() {
		if err := a.Run(); err != nil {
			log.Fatalf("Service running error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	a.Stop()

	log.Println("service is down")
}
