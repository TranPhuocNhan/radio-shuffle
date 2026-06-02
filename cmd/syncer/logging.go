package main

import (
	"log/slog"
	"os"
)

func setupLogger() {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))
}

func logAckError(message string, meta deliveryMeta, err error, extra []any) {
	attrs := meta.logAttrs(extra...)
	attrs = append(attrs, "err", err)
	slog.Error(message, attrs...)
}

func logDeliverySuccess(message string, meta deliveryMeta, extra []any) {
	slog.Info(message, meta.logAttrs(extra...)...)
}
