package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	collectorpb "github.com/Gilfoyle3301/system-stats-daemon/api/pb"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/config"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/grpcserver"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	configpath string
	grpcport   string
)

func main() {
	pflag.StringVar(&grpcport, "grpcport", "5005", "Port on which gRPC server will run")
	pflag.StringVar(&configpath, "config", "config.yml", "Path to configuration file")

	pflag.Parse()

	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Getwd failed: %v", err)
	}
	getParams, err := config.LoadConf(filepath.Join(currentDir, configpath))
	if err != nil {
		log.Fatalf("failed read config: %v", err)
	}

	go grpcserver.StartServer(&collector.Collector{}, grpcport, getParams)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%v", grpcport), opts...)
	if err != nil {
		log.Fatalf("connection failed: %v", err)
	}
	defer conn.Close()

	daemonClient := collectorpb.NewMetricsCollectorClient(conn)
	req := collectorpb.MetricsRequest{
		NSecond: int32(getParams.Interval),
		MSecond: 60,
	}

	stream, err := daemonClient.CollectMetrics(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error when calling CollectMetrics: %v", err)
	}
	for {

		resp, err := stream.Recv()
		if err != nil {
			log.Printf("Error when receiving a response: %v", err)
			break
		}
		log.Printf("Get metrics: %+v", resp.GetCollector())
	}

}
