package middleware

import (
	"context"

	clientnet "github.com/k1nky/ypmetrics/internal/client/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// XRealIPStreamInterceptor добавляет адрес хоста в мета-данные с ключом x-real-ip для запросов в потоке.
func XRealIPStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ip, _ := clientnet.RetriveClientAddress()
		newCtx := metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip.String())
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

// XRealIPUnaryInterceptor добавляет адрес хоста в мета-данные с ключом x-real-ip для унарных запросов.
func XRealIPUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ip, _ := clientnet.RetriveClientAddress()
		newCtx := metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip.String())
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}
