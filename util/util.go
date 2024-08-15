package util

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log/slog"
	"math/rand"
	"sync"
	"testing"
	"time"
)

const (
	// ConsistentRouterType 一致性哈希
	ConsistentRouterType = int32(iota)
	// RandomRouterType 随机路由
	RandomRouterType
	// WeightedRouterType 基于权重
	WeightedRouterType
	// TargetRouterType 特定路由
	TargetRouterType
	// KVRouterType 键值对路由
	KVRouterType
)

func Address(host string, port int32) string {
	return fmt.Sprintf("%s:%v", host, port)
}

func DicMap[K0 comparable, V0 any, K1 comparable, V1 any](
	m map[K0]V0, fun func(key K0, val V0) (K1, V1)) map[K1]V1 {
	rt := make(map[K1]V1)
	for k, v := range m {
		k1, v1 := fun(k, v)
		rt[k1] = v1
	}
	return rt
}

func SliceMap[T any, V any](list []T, fun func(T) V) []V {
	rt := make([]V, 0, len(list))
	for _, v := range list {
		rt = append(rt, fun(v))
	}
	return rt
}

func RandomDicValue[K comparable, V any](dict map[K]V) V {
	var rt V
	if len(dict) == 0 {
		return rt
	}
	index := rand.Intn(len(dict))
	count := 0
	for _, value := range dict {
		if count < index {
			count += 1
			continue
		}
		rt = value
		break
	}
	return rt
}

func RandomItem[T any](items []T) T {
	var rt T
	if len(items) == 0 {
		return rt
	}
	index := rand.Intn(len(items))
	return items[index]
}

func Info(format string, a ...any) {
	slog.Info(fmt.Sprintf(format, a...))
}

func Error(format string, a ...any) {
	slog.Error(fmt.Sprintf(format, a...))
}

func Warn(format string, a ...any) {
	slog.Warn(fmt.Sprintf(format, a...))
}

type ServiceInfo struct {
	Name      string
	Instances []DstInstanceInfo
}

func (s *ServiceInfo) GetServiceName() string {
	return s.Name
}
func (s *ServiceInfo) GetInstance() []DstInstanceInfo {
	return s.Instances
}

type DstService interface {
	GetServiceName() string
	GetInstance() []DstInstanceInfo
}

type DstInstanceInfo interface {
	GetInstanceID() int64
	GetName() string
	GetWeight() int32
	GetHost() string
	GetPort() int32
}

type SyncContainer[T any] struct {
	Mutex sync.Mutex
	Value T
}

func RandomIP() string {
	return fmt.Sprintf("%v.%v.%v.%v",
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
		rand.Intn(256),
	)
}

func RandomPort() int32 {
	return rand.Int31n(65535)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func ClearEtcd(address string, t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{address},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = client.Delete(ctx, "", clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}
}
