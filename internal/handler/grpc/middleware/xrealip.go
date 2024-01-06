package middleware

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrUntrustedSourceAddress = status.Error(codes.PermissionDenied, "the request is not within trusted network")
)

func resolveIP(ctx context.Context) net.IP {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	ips := md.Get("x-real-ip")
	if len(ips) == 0 {
		return nil
	}
	return net.ParseIP(ips[0])
}

func XRealIPStreamInterceptor(subnet net.IPNet) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ip := resolveIP(ss.Context())
		if ip == nil {
			return ErrUntrustedSourceAddress
		}
		if !subnet.Contains(ip) {
			return ErrUntrustedSourceAddress
		}
		return handler(srv, ss)
	}
}

func XRealIPUnaryInterceptor(subnet net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ip := resolveIP(ctx)
		if ip == nil {
			return nil, ErrUntrustedSourceAddress
		}
		if !subnet.Contains(ip) {
			return nil, ErrUntrustedSourceAddress
		}
		return handler(ctx, req)
	}
}
