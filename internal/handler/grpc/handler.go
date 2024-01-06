// Пакет grpc реализует обработчик gRPC запросов к серверу сбора метрик
package grpc

import (
	"context"
	"io"

	"github.com/k1nky/ypmetrics/internal/entities/metric"
	pb "github.com/k1nky/ypmetrics/internal/proto"
	"github.com/k1nky/ypmetrics/internal/usecases/keeper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler обработчик grpc-запросов.
type Handler struct {
	pb.UnimplementedMetricsServer
	keeper *keeper.Keeper
}

// New возвращает новый обработчик grpc-запросов к серверу метрик
func New(keeper *keeper.Keeper) *Handler {
	return &Handler{
		keeper: keeper,
	}
}

// UpdateMetric обновление одной метрики.
func (h *Handler) UpdateMetric(ctx context.Context, in *pb.UpdateMetricRequest) (*pb.UpdateMetricsResponse, error) {
	switch in.Type {
	case pb.Type_COUNTER:
		if err := h.keeper.UpdateCounter(ctx, in.Name, in.Delta); err != nil {
			return nil, status.Errorf(codes.Internal, "update counter: %s", err)
		}
	case pb.Type_GAUGE:
		if err := h.keeper.UpdateGauge(ctx, in.Name, in.Value); err != nil {
			return nil, status.Errorf(codes.Internal, "update gauge: %s", err)
		}
	}

	return &pb.UpdateMetricsResponse{}, nil
}

// UpdateMetrics обновление нескольких метрик.
func (h *Handler) UpdateMetrics(stream pb.Metrics_UpdateMetricsServer) error {
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.DataLoss, "update metrics: %s", err)
		}
		metrics := metric.NewMetrics()
		for _, v := range request.Mertics {
			switch v.Type {
			case pb.Type_COUNTER:
				metrics.Counters = append(metrics.Counters, metric.NewCounter(v.Name, v.Delta))
			case pb.Type_GAUGE:
				metrics.Gauges = append(metrics.Gauges, metric.NewGauge(v.Name, v.Value))
			}
		}
		if err := h.keeper.UpdateMetrics(stream.Context(), *metrics); err != nil {
			return status.Errorf(codes.Internal, "failed to update metrics %s", err)
		}
	}
	return nil
}
