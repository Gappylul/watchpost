package grpcserver

import (
	"context"
	"time"

	"github.com/gappylul/watchpost/internal/hub"
	pb "github.com/gappylul/watchpost/proto"
)

type Server struct {
	pb.UnimplementedWatchpostServer
	hub *hub.Hub
}

func New(h *hub.Hub) *Server {
	return &Server{hub: h}
}

func (s *Server) GetServices(ctx context.Context, req *pb.GetServicesRequest) (*pb.GetServicesResponse, error) {
	statuses := s.hub.Latest()
	services := make([]*pb.Service, 0, len(statuses))
	for _, st := range statuses {
		services = append(services, toProto(st))
	}
	return &pb.GetServicesResponse{Services: services}, nil
}

func (s *Server) WatchStatus(req *pb.WatchStatusRequest, stream pb.Watchpost_WatchStatusServer) error {
	ch := s.hub.Subscribe()
	defer s.hub.Unsubscribe(ch)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case st := <-ch:
			if err := stream.Send(toProto(st)); err != nil {
				return err
			}
		}
	}
}

func toProto(s hub.Status) *pb.Service {
	return &pb.Service{
		Name:      s.Service,
		Status:    s.Status,
		LatencyMs: s.LatencyMs,
		Error:     s.Error,
		CheckedAt: s.CheckedAt.Format(time.RFC3339),
	}
}
