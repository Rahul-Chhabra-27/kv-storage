package main
import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
	"fmt"
	"gorm.io/gorm"
)
func (KvServerManager *KvService) DeleteKeyValue(ctx context.Context,request *kvpb.DeleteKeyValueRequest) (*kvpb.DeleteKeyValueResponse, error) {
	key := request.Key

	// Check if key is missing
	if key == "" {
		logger.Info("Key missing in delete request")
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Key missing in delete request",
			StatusCode: int64(StatusBadRequest),
		}, nil
	}

	// Check if key exists in DB
	var existingKeyValuePair model.KV
	err := kvDbConnector.Where("key_name = ?", key).First(&existingKeyValuePair).Error

	if err == gorm.ErrRecordNotFound {
		logger.Info(fmt.Sprintf("Key %s not found for deletion", key))
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Key not found",
			StatusCode: int64(StatusNotFound),
		}, nil
	} else if err != nil {
		logger.Error(fmt.Sprintf("Database error while deleting key %s: %v", key, err))
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Database error",
			StatusCode: int64(StatusInternalServerError),
		}, nil
	}

	// Delete key-value pair
	deleteResult := kvDbConnector.Delete(&existingKeyValuePair)
	if deleteResult.Error != nil {
		logger.Error(fmt.Sprintf("Failed to delete key %s: %v", key, deleteResult.Error))
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Failed to delete key-value pair",
			StatusCode: int64(StatusInternalServerError),
		}, nil
	}

	logger.Info(fmt.Sprintf("Key-Value Pair %s deleted successfully", key))
	return &kvpb.DeleteKeyValueResponse{
		Message:    "Key-value pair successfully deleted",
		StatusCode: int64(StatusOK),
	}, nil
}