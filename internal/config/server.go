package config

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	// addr - только loopback: снаружи запросы приходят через Caddy на том же сервере, напрямую порт недоступен
	addr string = `127.0.0.1:8000`
	// debugAddr - для локальной разработки все интерфейсы: сайт открывается с телефона по IP компьютера
	debugAddr string = `:8000`
	// shutdownTimeout - сколько при остановке ждать начатые запросы. docker stop через 10 секунд убивает процесс
	shutdownTimeout time.Duration = time.Second * 5
)

// Debug - запуск для локальной разработки: задана переменная окружения DEBUG, значение не важно
func Debug() bool {
	_, ok := os.LookupEnv("DEBUG")

	return ok
}

// NewServer - собирает http-сервер, маршруты регистрирует routes. debug — см. Debug
func NewServer(routes func(chi.Router), debug bool) *http.Server {
	log := slog.Default()

	router := chi.NewRouter()

	// id запроса нужен журналу: по нему ошибка обработчика находится рядом со строкой самого запроса
	router.Use(middleware.RequestID)
	router.Use(requestLogger(log))
	router.Use(recoverer(log))

	routes(router)

	listen := addr
	if debug {
		listen = debugAddr
	}

	return &http.Server{
		Addr:              listen,
		Handler:           router,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
		ReadTimeout:       time.Second * 5,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 5,
		IdleTimeout:       time.Second * 5,
	}
}

// Serve - принимает запросы, пока не отменят ctx, и затем мягко останавливает сервер:
// новые соединения больше не принимаются, а начатые запросы дорабатывают
func Serve(ctx context.Context, server *http.Server) error {
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("can't listen on %s: %w", server.Addr, err)
	}

	return serve(ctx, server, ln)
}

func serve(ctx context.Context, server *http.Server, ln net.Listener) error {
	served := make(chan error, 1)
	go func() {
		served <- server.Serve(ln)
	}()

	select {
	case err := <-served:
		return fmt.Errorf("server stopped: %w", err)
	case <-ctx.Done():
	}

	slog.InfoContext(ctx, "shutting down server")

	// ctx уже отменён, а начатым запросам нужно время, чтобы доработать
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("can't shut down server gracefully: %w", err)
	}

	return nil
}
