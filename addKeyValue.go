package main

import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
	"fmt"
	"gorm.io/gorm"
)
func (KvServerManager *KvService) SetKeyValue(ctx context.Context, request *kvpb.SetKeyValueRequest) (*kvpb.SetKeyValueResponse, error) {

	 key := request.Key;
	 value := request.Value;

	// check if the key or value is missing!
	if key == "" || value == ""  {
		logger.Info("Either key or value missing")
		return &kvpb.SetKeyValueResponse{
			Message: "Either key or value missing",
			StatusCode : int64(StatusBadRequest),
		}, nil
	}
		
		// Check if it's already present in the db
		var existingKeyValuePair model.KV;
		KeyNotFoundError := kvDbConnector.Where("key_name = ?", key).First(&existingKeyValuePair).Error

		if KeyNotFoundError == gorm.ErrRecordNotFound {
			
			newKVPair := &model.KV{ Key : key, Value : value}
			// It will return a primary key in the result.
			kvDbConnector.Create(newKVPair);
			// Add it into the cacahe!
			cache.Put(key,value);

			logger.Info(fmt.Sprintf("Key-Value Pair %s created successfully with primary key %s", newKVPair.Key,newKVPair.ID))

			return &kvpb.SetKeyValueResponse{
				Message: "Key-value pair successfully created",
				StatusCode : int64(StatusCreated),
			}, nil
	} else if KeyNotFoundError == nil {
		// Key exists — update its value
		// updateResult := kvDbConnector.Model(&existingKeyValuePair).Update("value", value)
		// if updateResult.Error != nil {
		// 	logger.Error(fmt.Sprintf("Error updating KV pair: %v", updateResult.Error))
		// 	return &kvpb.SetKeyValueResponse{
		// 		Message:    "Failed to update key-value pair",
		// 		StatusCode: int64(StatusInternalServerError),
		// 	}, nil
		// }

		// logger.Info(fmt.Sprintf("Key-Value Pair %s updated successfully", key))
		// return &kvpb.SetKeyValueResponse{
		// 		Message:    "Key-value pair successfully updated",
		// 		StatusCode: int64(StatusOK),
		// 	}, nil
		logger.Info(fmt.Sprintf("Key-Value Pair %s already present", key))
		return &kvpb.SetKeyValueResponse{
				Message:    "Key-value pair already present",
				StatusCode: int64(StatusConflict),
		}, nil
	} else {
		logger.Error(fmt.Sprintf("Database error: %v", KeyNotFoundError))
		return &kvpb.SetKeyValueResponse{
			Message:    "Database error",
			StatusCode: int64(StatusInternalServerError),
		}, nil
	}
}