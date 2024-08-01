package collector_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	collectorpb "github.com/Gilfoyle3301/system-stats-daemon/api/pb"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/collector"
	"github.com/Gilfoyle3301/system-stats-daemon/internal/grpcserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerIntegration(t *testing.T) {
	done := make(chan struct{})
	ticker := time.NewTicker(3 * time.Second)
	var lastResp *collectorpb.MetricsResponse
	go grpcserver.StartServer(&collector.Collector{}, "12345")
	time.Sleep(2 * time.Second)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(fmt.Sprintf("localhost:%v", "12345"), opts...)
	require.NoError(t, err)
	defer conn.Close()

	daemonClient := collectorpb.NewMetricsCollectorClient(conn)
	req := collectorpb.MetricsRequest{
		NSecond: 2,
	}

	stream, err := daemonClient.CollectMetrics(context.Background(), &req)
	require.NoError(t, err)
	go func() {
		select {
		case <-ticker.C:
			resp, err := stream.Recv()
			require.NoError(t, err)
			lastResp = resp
		case <-done:
			return
		}
	}()
	go func() {
		time.Sleep(10 * time.Second)
		done <- struct{}{}
	}()

	select {
	case <-time.After(30 * time.Second):
		close(done)
	case <-done:
	}
	assert.NotNil(t, lastResp, "Expected to receive awt least one metrics response")
}
