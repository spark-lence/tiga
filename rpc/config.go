package rpc

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/spark-lence/tiga"
	pb "github.com/spark-lence/tiga/rpc/pb"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type ConfigServer struct {
	pb.UnimplementedConfigServer
	configs     map[string]*tiga.Configuration
	settingsDir string
}

func (s *ConfigServer) GetConfig(ctx context.Context, in *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	config := s.configs[in.Env]
	val := config.Get(in.Key)
	if val == nil {
		return nil, fmt.Errorf("Not found config key:%s", in.Key)
	}
	value, err := json.Marshal(val)

	return &pb.ConfigResponse{Value: value}, err
}
func (s *ConfigServer) SetConfig(ctx context.Context, in *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	config := s.configs[in.Env]
	config.SetConfig(in.Key, in.Value, in.Env)

	return nil, nil
}
func NewConfigServer(settingDir string) *ConfigServer {
	configs := make(map[string]*tiga.Configuration)
	configs["dev"] = tiga.InitSettings("dev", settingDir)
	configs["prd"] = tiga.InitSettings("prd", settingDir)
	return &ConfigServer{
		configs:     configs,
		settingsDir: settingDir,
	}
}
func (s *ConfigServer) Start() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterConfigServer(server, s)
	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *ConfigServer) Refrsh() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	pb.RegisterConfigServer(server, s)
	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
