package server

import (
	"context"
	"fmt"

	pb "github.com/muyi-zcy/tech-muyi-base-go/example/proto/echo"
	"github.com/muyi-zcy/tech-muyi-base-go/myLogger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type echoServer struct {
	pb.UnimplementedEchoServiceServer
}

func RegisterEchoService(s *grpc.Server) {
	pb.RegisterEchoServiceServer(s, &echoServer{})
}

func (s *echoServer) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	return &pb.PingReply{Message: "pong from example-producer"}, nil
}

func (s *echoServer) Echo(_ context.Context, req *pb.EchoRequest) (*pb.EchoReply, error) {
	msg := req.GetMessage()
	myLogger.Info("收到 Echo 请求", zap.String("message", msg))
	return &pb.EchoReply{Message: fmt.Sprintf("producer echo: %s", msg)}, nil
}
