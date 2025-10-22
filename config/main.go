package config
import (
	"context"
	"fmt"
	"google.golang.org/grpc"
)

func UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	fmt.Println("--> UnaryInterceptor: ", info.FullMethod)
	return handler(ctx, req)
}