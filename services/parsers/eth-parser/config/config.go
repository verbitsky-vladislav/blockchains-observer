package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Network string

const (
	ETH   Network = "ETH"
	MATIC Network = "MATIC"
	BSC   Network = "BSC"
)

type Stage string

const (
	Development Stage = "development"
	Staging     Stage = "staging"
	Production  Stage = "production"
)

type Config struct {
	ServiceName string
	Stage       Stage

	Network           Network
	WsEndpoints       []string
	HttpEndpoint      string
	RetryCount        int
	ReconnectInterval time.Duration
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	network, err := parseNetwork(os.Getenv("NETWORK"))
	if err != nil {
		return nil, err
	}

	stage, err := parseStage(os.Getenv("STAGE"))
	if err != nil {
		return nil, err
	}

	// мы храним урлы эндпоинтов через запятую
	wsUrlsEnv := os.Getenv("WS_ENDPOINTS")
	wsEndpoints := strings.Split(wsUrlsEnv, ",")
	for i, url := range wsEndpoints {
		wsEndpoints[i] = strings.TrimSpace(url)
	}

	retryCount, err := strconv.Atoi(os.Getenv("RETRY_CONNECTION_COUNT"))
	if err != nil {
		return nil, fmt.Errorf("invalid RETRY_CONNECTION_COUNT: %v", err)
	}

	reconnectInterval, err := parseReconnectInterval(os.Getenv("RECONNECT_INTERVAL"))
	if err != nil {
		return nil, fmt.Errorf("invalid RECONNECT_INTERVAL: %v", err)
	}

	return &Config{
		ServiceName:       os.Getenv("SERVICE_NAME"),
		Stage:             stage,
		Network:           network,
		WsEndpoints:       wsEndpoints,
		HttpEndpoint:      os.Getenv("HTTP_ENDPOINT"),
		RetryCount:        retryCount,
		ReconnectInterval: reconnectInterval, // here is in second
	}, nil
}

func parseReconnectInterval(value string) (time.Duration, error) {
	if value == "" {
		return 0, fmt.Errorf("RECONNECT_INTERVAL is empty") // here is must be custom error to correctly catch it upper
	}

	seconds, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds) * time.Second, nil
}

func parseStage(value string) (Stage, error) {
	switch Stage(value) {
	case Development, Staging, Production:
		return Stage(value), nil
	default:
		return "", fmt.Errorf("unsupported stage value in env: %s", value)
	}
}

func parseNetwork(value string) (Network, error) {
	switch Network(value) {
	case ETH, MATIC, BSC:
		return Network(value), nil
	default:
		return "", fmt.Errorf("unsupported network value in env: %s", value)
	}
}
