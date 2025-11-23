package main

import (
	"context"
	kvpb "github.com/kv-storage/proto/kv"
	"github.com/kv-storage/model"
	"fmt"
	"strings"
)
func (KvServerManager *KvService) SetKeyValue(
    ctx context.Context,
    request *kvpb.SetKeyValueRequest,
) (*kvpb.SetKeyValueResponse, error) {

    key := request.Key
    value := request.Value

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

        logger.Error(fmt.Sprintf("DB insert error: %v", err))
        return &kvpb.SetKeyValueResponse{
            Message:    "Database error",
            StatusCode: int64(StatusInternalServerError),
        }, nil
    }
    cache.Put(key, value)
    return &kvpb.SetKeyValueResponse{
        Message:    "Key-value pair successfully created",
        StatusCode: int64(StatusCreated),
    }, nil
}
