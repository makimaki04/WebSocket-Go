package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/makimaki04/WebSocket-Go/internal/wsserver"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log := logger.Sugar()
	wsSrv := wsserver.NewWsServer(":8080", log)
	log.Info("ws server succesfully started")

	errCh := make(chan error, 1)
	go func() {
		errCh <- wsSrv.Start()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("ws server start: %v", err)
		}
		return
	case <-ctx.Done():
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := wsSrv.Stop(stopCtx); err != nil {
		log.Errorf("ws server stop: %v", err)
	}

	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Errorf("ws server start: %v", err)
	}
}
