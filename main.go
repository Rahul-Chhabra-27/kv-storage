package main
import (
	"log"
	"net"
	"context"
	"net/http"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/kv-storage/config"
)

const (
	StatusBadRequest       = 400
	StatusConflict         = 409
	StatusInternalServerError = 500
	StatusOK               = 200
	StatusCreated          = 201
	StatusNotFound         = 404
	StatusUnauthorized     = 401
	StatusForbidden        = 403
)
var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
}

// Responsible for starting the server
func startServer() {
	// flush logger buffer on exit
	defer logger.Sync()

	// Log a message
	logger.Info("Starting server...")
	
	// Initialize the gotenv file..
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file", zap.Error(err))
	}

	// Creating TCP Socket listener on port 50051
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(config.UnaryInterceptor),
	)

	// Start the server in a new goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			logger.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	// Create a new gRPC-Gateway server
	// it connect to the gRPC server we just started and act as a grpc client!
	_, err = grpc.DialContext(
		context.Background(),
		"localhost:50051",
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("Failed to dial server", zap.Error(err))
	}
	// Create a new gRPC-Gateway mux
	gwmux := runtime.NewServeMux()

	// Create a new HTTP server
	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}
	logger.Info("Serving gRPC-Gateway", zap.String("address", "http://0.0.0.0:8090"))
	if err := gwServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Failed to listen and serve: %v", err)
	}
	
}

func main() {
	// Start the server
	startServer()
}