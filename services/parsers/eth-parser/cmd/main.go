package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/verbitsky-vladislav/blockchains-observer/pkg/logger"
	"github.com/verbitsky-vladislav/blockchains-observer/services/parsers/eth-parser/config"
	"github.com/verbitsky-vladislav/blockchains-observer/services/parsers/eth-parser/internal/blockchain"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Errorf("error loading config in eth-parser microservice: %v", err))
	}

	logg, err := logger.NewLogger(cfg.ServiceName)
	if err != nil {
		panic(fmt.Errorf("error creating logger: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	monitor := blockchain.NewMonitoring(cfg, ctx, logg)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		logg.Info("starting monitoring...", "main.go")
		monitor.Init()
		logg.Info("monitoring stopped", "main.go")
	}()

	if err := http.ListenAndServe(":8080", nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logg.Panic("error starting HTTP server: %v", err)
	}

	//handleGracefulShutdown(ctx, cancel, logg, &wg)

	//logg.Info("application exited gracefully.", "main.go")
}

// handleGracefulShutdown управляет завершением работы приложения
func handleGracefulShutdown(ctx context.Context, cancel context.CancelFunc, logg *logger.Logger, wg *sync.WaitGroup) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		logg.Info("Received signal: %v", map[string]interface{}{"signal": sig})

		cancel()

		logg.Info("waiting for all routines to finish...", "main.go")
		wg.Wait()
		logg.Info("all routines finished", "main.go")
		os.Exit(0)
	}()
}
