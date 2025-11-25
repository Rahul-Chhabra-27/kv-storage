package main

import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
	"gorm.io/gorm"
)

func (KvServerManager *KvService) DeleteKeyValue(ctx context.Context, request *kvpb.DeleteKeyValueRequest) (*kvpb.DeleteKeyValueResponse, error) {
	key := request.Key

	// Check if key is missing
	if key == "" {
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Key missing in delete request",
			StatusCode: int64(StatusBadRequest),
		}, nil
	}

	// Check if key exists in DB
	var existingKeyValuePair model.KV
	err := kvDbConnector.Where("key_name = ?", key).First(&existingKeyValuePair).Error

	if err == gorm.ErrRecordNotFound {
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Key not found",
			StatusCode: int64(StatusNotFound),
		}, nil
	} else if err != nil {
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Database error",
			StatusCode: int64(StatusInternalServerError),
		}, nil
	}

	// Delete key-value pair
	deleteResult := kvDbConnector.Delete(&existingKeyValuePair)
	if deleteResult.Error != nil {
		return &kvpb.DeleteKeyValueResponse{
			Message:    "Failed to delete key-value pair",
			StatusCode: int64(StatusInternalServerError),
		}, nil
	}
	cache.DeleteKey(key)
	return &kvpb.DeleteKeyValueResponse{
		Message:    "Key-value pair successfully deleted",
		StatusCode: int64(StatusOK),
	}, nil
}