package sdk

import (
	"fmt"
	"log/slog"
	"math/rand"
)

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
