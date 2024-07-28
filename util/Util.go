package util

import (
	"Supernove2024/miniRouterProto"
	"fmt"
	"github.com/go-redsync/redsync/v4"
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

func TryUnlock(mutex *redsync.Mutex) {
	_, err := mutex.Unlock()
	if err != nil {
		Error("释放redis锁失败, Name:%s, err:%v", mutex.Name(), err)
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
