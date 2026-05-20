package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Gappylul/goverseer"
	"github.com/gappylul/watchpost/internal/api"
	"github.com/gappylul/watchpost/internal/checker"
	"github.com/gappylul/watchpost/internal/config"
	"github.com/gappylul/watchpost/internal/grpcserver"
	"github.com/gappylul/watchpost/internal/hub"
	pb "github.com/gappylul/watchpost/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	configPath := flag.String("config", "watchpost.yml", "path to config file")
	addr := flag.String("addr", ":8080", "http server address")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	h := hub.New()

	children := make([]goverseer.ChildSpec, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		var start func(ctx context.Context) error
		switch svc.Check {
		case config.HTTP:
			start = checker.NewHTTPChecker(svc, h)
		case config.TCP:
			start = checker.NewTCPChecker(svc, h)
		default:
			log.Fatalf("unknown check type: %s", svc.Check)
		}
		children = append(children, goverseer.ChildSpec{
			Name:    svc.Name,
			Start:   start,
			Restart: goverseer.Permanent,
		})
	}

	sup := goverseer.New(
		goverseer.OneForOne,
		goverseer.WithName("watchpost-supervisor"),
		goverseer.WithIntensity(5, time.Minute),
		goverseer.WithChildren(children...),
	)

	if err = sup.Start(); err != nil {
		log.Fatalf("starting supervisor: %v", err)
	}

	handler := api.New(h)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("watchpost listening on %s", *addr)

	grpcAddr := flag.String("grpc-addr", ":9090", "grpc server address")

	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWatchpostServer(grpcServer, grpcserver.New(h))
	reflection.Register(grpcServer)

	go func() {
		log.Printf("grpc listening on %s", *grpcAddr)
		if err = grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	if err = http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
