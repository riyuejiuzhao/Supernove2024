package grpc_sdk

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc/metadata"
)

type MetaData struct {
	*metadata.MD
	nowRouterKey map[string]string
}

func NewMetaData() *MetaData {
	return &MetaData{
		MD:           &metadata.MD{},
		nowRouterKey: make(map[string]string),
	}
}

func (m *MetaData) SetRouterType(routerType string) {
	m.Set(RouterTypeHeader, routerType)
}

func (m *MetaData) AddRouterTag(key string, value string) error {
	m.nowRouterKey[key] = value
	js, err := json.Marshal(m.nowRouterKey)
	if err != nil {
		return err
	}
	m.Set(RouterKeyHeader, string(js))
	return nil
}

func (m *MetaData) NewOutgoingContext(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, *m.MD)
}
