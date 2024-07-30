package util

import (
	"Supernove2024/miniRouterProto"
	"fmt"
	"log/slog"
	"math/rand"
)

func NewServiceInfo(serviceName string) *miniRouterProto.ServiceInfo {
	return &miniRouterProto.ServiceInfo{
		ServiceName: serviceName,
		Revision:    0,
		Instances:   make([]*miniRouterProto.InstanceInfo, 0),
	}
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
