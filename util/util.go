package util

import (
	"fmt"
	"log/slog"
	"math/rand"
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
