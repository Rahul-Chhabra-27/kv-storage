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
		// TODO : Check if this already present in cache or not.

		// Check if it's already present in the db
		var existingKeyValuePair model.KV;
		KeyNotFoundError := kvDbConnector.Where("key_name = ?", key).First(&existingKeyValuePair).Error

		if KeyNotFoundError == gorm.ErrRecordNotFound {
			
			newKVPair := &model.KV{ Key : key, Value : value}
			// It will return a primary key in the result.
			primaryKey := kvDbConnector.Create(newKVPair);

			logger.Info(fmt.Sprintf("Key-Value Pair %s created successfully with primary key %s", newKVPair.Key,primaryKey))

			return &kvpb.SetKeyValueResponse{
				Message: "Key-value pair successfully created",
				StatusCode : int64(StatusCreated),
			}, nil
	}
	logger.Info(fmt.Sprintf("Key-Value Pair Already Exist!"))
	return &kvpb.SetKeyValueResponse{
			Message: "KV Pair Already Exist",
			StatusCode : int64(StatusConflict),
	}, nil
}