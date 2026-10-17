// Package logger - структурные логи приложения на log/slog
package logger

import (
	"log/slog"
	"os"
)

// Setup - логгер по умолчанию для всего приложения, в том числе для стандартного log.
// В продакшене JSON по записи на строку с уровнем info: такие логи фильтруются по полям.
// При отладке (debug) — читаемый текст с уровнем debug
func Setup(debug bool) {
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, nil)
	if debug {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	slog.SetDefault(slog.New(handler))
}

// Fatal - ошибка, после которой работать нельзя (обычно при запуске): пишет её и завершает процесс
func Fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}
