package config
import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"os"
	"log"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/kv-storage/model"
	"time"
)
func DatabaseDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)
}
func GoDotEnvVariable(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func ConnectDB() (*gorm.DB, error) {
	// Responsible for connecting to the database
	kvdb, err := gorm.Open(mysql.Open(DatabaseDsn()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Set connection pool limits
	sqlDB, err := kvdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(2000)              // Max active connections
	sqlDB.SetMaxIdleConns(100)              // Idle connections to keep
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // Recycle connections
	// Migrate the schema
	kvdb.AutoMigrate(&model.KV{})
	return kvdb, nil
}
func UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	return handler(ctx, req)
}