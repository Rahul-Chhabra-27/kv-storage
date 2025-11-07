package main

import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
)

func (KvServerManager *KvService) GetKeyValue(ctx context.Context, request *kvpb.GetKVRequest) (*kvpb.GetKVResponse, error) {
	// getting the key from request...
	key := request.Key;
	
	// checking in the cache
	value,isValueExist := cache.Get(key);
	if isValueExist == true  {
		logger.Info("Response is Coming from the Cache (CPU) Operation!");
		return &kvpb.GetKVResponse{
			Message:"Key found",
			StatusCode : int64(StatusOK),
			Value:value,
		},nil
	}
	// Checking into the database
	var keyValue model.KV
	if err := kvDbConnector.Where("key_name = ?", key).First(&keyValue).Error; err != nil {
		logger.Error("Key not found")
		return &kvpb.GetKVResponse{
			Message:    "Key not found",
			StatusCode: int64(StatusNotFound),
		}, nil
	}
	if(!isValueExist) {
		cache.Put(key,keyValue.Value);
	}
	logger.Info("Response is coming from the DB (I/O) Operation!");
	return &kvpb.GetKVResponse{
		Message:"Key found",
		StatusCode : int64(StatusOK),
		Value:keyValue.Value,
	},nil
}
