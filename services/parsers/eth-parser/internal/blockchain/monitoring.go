package blockchain

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/verbitsky-vladislav/blockchains-observer/pkg/logger"
	"github.com/verbitsky-vladislav/blockchains-observer/services/parsers/eth-parser/config"
	"time"
)

type Monitoring struct {
	cfg  *config.Config
	ctx  context.Context
	logg *logger.Logger
}

func NewMonitoring(cfg *config.Config, ctx context.Context, logg *logger.Logger) *Monitoring {
	return &Monitoring{
		cfg:  cfg,
		ctx:  ctx,
		logg: logg,
	}
}

func (m *Monitoring) Init() {
	// делаем попытку восстановиться, если какая-то ошибка
	defer func() {
		if r := recover(); r != nil {
			m.logg.Error("unexpected panic in monitoring init: %v", map[string]interface{}{"error": r})
		}
	}()

	// создаем контекст, чтобы закрыть соединения и подписку на новые хедеры блоков
	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	// подключаемся к Ethereum через WebSocket
	client, err := m.createConnection(ctx)
	if err != nil {
		m.logg.Error("failed to create connection: %v", err)
		return
	}

	// канал для получения заголовков
	headers := make(chan *types.Header)
	defer close(headers) // не забываем закрыть канал

	// создание самого соединения
	sub, err := m.subscribeToNewHead(ctx, client, headers)
	if err != nil {
		m.logg.Error("failed to subscribe to new blocks: %v", err)
		return
	}
	defer sub.Unsubscribe()

	// в бесконечном цикле обрабатываем возможные варианты заголовков
	for {
		select {
		// case 1 : если сервис закончил работу, то мы прекращаем бесконечный цикл и выходим из него
		// после выхода отработают все defer : unsub && close connection
		case <-ctx.Done():
			m.logg.Info("service is off with graceful shutdown", nil)
			return
		// case 2 : если в подписке идет ошибка, то мы пробуем переподключиться к соединению
		// так же, мы закрываем текущую горутину, так как выходим из цикла
		case subErr := <-sub.Err():
			m.logg.Error("subscription error: %v", subErr)
			// подключаемся к сокету с пересозданием клиента
			client, err = m.reconnect(ctx)
			if err != nil {
				m.logg.Error("failed to reconnect: %v", err)
				return
			}

			// обновляем подписку на новые блоки
			sub, err = m.subscribeToNewHead(ctx, client, headers)
			if err != nil {
				m.logg.Error("failed to resubscribe: %v", err)
				return
			}
		// case 3 : обработка блока
		// todo add block processing method
		case header := <-headers:
			// получение блока и его данных
			block, blockErr := client.BlockByHash(ctx, header.Hash())
			if blockErr != nil {
				m.logg.Error("failed to fetch block: %v", blockErr)
				continue
			}
			// mock block processing
			m.logg.Info("block received", map[string]interface{}{
				"hash":         block.Hash().Hex(),
				"blockNumber":  block.Number(),
				"transactions": len(block.Transactions()),
			})
		}
	}
}

// createConnection - подключаемся к доступному вебсокету, пробегаясь по общему массиву
func (m *Monitoring) createConnection(ctx context.Context) (*ethclient.Client, error) {
	for _, endpoint := range m.cfg.WsEndpoints {
		client, err := m.tryConnect(ctx, endpoint)
		if err == nil {
			return client, nil
		}
		m.logg.Error("failed to connect to endpoint", map[string]interface{}{
			"endpoint": endpoint,
			"error":    err,
		})
	}
	return nil, fmt.Errorf("failed to connect to any WebSocket endpoint after multiple attempts")
}

// tryConnect - попытка подключиться к конкретному эндпоинту
func (m *Monitoring) tryConnect(ctx context.Context, endpoint string) (*ethclient.Client, error) {
	for attempt := 0; attempt < m.cfg.RetryCount; attempt++ {
		client, err := ethclient.DialContext(ctx, endpoint)
		if err == nil {
			return client, nil
		}
		m.logg.Error("connection attempt failed", map[string]interface{}{
			"endpoint": endpoint,
			"attempt":  attempt + 1,
			"error":    err,
		})
		if attempt < m.cfg.RetryCount-1 {
			time.Sleep(m.cfg.ReconnectInterval)
		}
	}
	return nil, fmt.Errorf("failed to connect to endpoint %s after %d attempts", endpoint, m.cfg.RetryCount)
}

// reconnect - обработка повторного подключения
func (m *Monitoring) reconnect(ctx context.Context) (*ethclient.Client, error) {
	client, err := m.createConnection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to reconnect after subscription error: %v", err)
	}
	return client, nil
}

// subscribeToNewHead - создание подписки на новые блоки
func (m *Monitoring) subscribeToNewHead(ctx context.Context, client *ethclient.Client, headers chan *types.Header) (ethereum.Subscription, error) {
	sub, err := client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new blocks: %v", err)
	}
	return sub, nil
}

func (m *Monitoring) getNextWsEndpoint(currentIndex int) (string, int) {
	nextIndex := (currentIndex + 1) % len(m.cfg.WsEndpoints)
	return m.cfg.WsEndpoints[nextIndex], nextIndex
}
