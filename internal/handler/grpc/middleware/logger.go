package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type requestLogger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

func LoggerStreamInterceptor(l requestLogger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ss.Context())
		err := handler(srv, ss)
		if err != nil {
			l.Infof("%s %v %v", info.FullMethod, md, time.Since(start))
		} else {
			l.Errorf("%s %v %v", info.FullMethod, md, err)
		}
		return err
	}
}

func LoggerUnaryInterceptor(l requestLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		md, _ := metadata.FromIncomingContext(ctx)
		resp, err = handler(ctx, req)
		if err == nil {
			l.Infof("%s %v %v", info.FullMethod, md, time.Since(start))
		} else {
			l.Errorf("%s %v %v", info.FullMethod, md, err)
		}
		return resp, err
	}
}
