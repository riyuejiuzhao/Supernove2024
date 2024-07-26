package util

import (
	"fmt"
	"log/slog"
	"math/rand"
)

func RandomDicValue[K comparable, V any](dict map[K]V) V {
	index := rand.Intn(len(dict))
	count := 0
	var rt V
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
