package main

import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
	"strings"
    // "log"
)
func (KvServerManager *KvService) SetKeyValue(
    ctx context.Context,
    request *kvpb.SetKeyValueRequest,
) (*kvpb.SetKeyValueResponse, error) {

    key := request.Key
    value := request.Value
    // log.Printf("Received SetKeyValue request - Key: %s, Value: %s", key, value)
    if key == "" || value == "" {
        return &kvpb.SetKeyValueResponse{
            Message:    "Either key or value missing",
            StatusCode: int64(StatusBadRequest),
        }, nil
    }

    kv := model.KV{Key: key, Value: value}
     
    if err := kvDbConnector.Create(&kv).Error; err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            return &kvpb.SetKeyValueResponse{
                Message:    "Key already exists",
                StatusCode: int64(StatusConflict),
            }, nil
        }
        return &kvpb.SetKeyValueResponse{
            Message:    "Database error",
            StatusCode: int64(StatusInternalServerError),
        }, nil
    }
    cache.Put(key, value)
    // log.Printf("Key-Value pair set successfully - Key: %s, Value: %s", key, value)
    return &kvpb.SetKeyValueResponse{
        Message:    "Key-value pair successfully created",
        StatusCode: int64(StatusCreated),
    }, nil
}
