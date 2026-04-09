package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/user/todo-api/config"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/infrastructure/push"
	"github.com/user/todo-api/internal/repository"
)

const (
	pollInterval     = 10 * time.Second
	batchSize        = 100
	maxRetries       = 3
	shutdownTimeout  = 30 * time.Second
)

type Worker struct {
	queueRepo      domain.NotificationQueueRepository
	tokenRepo      *repository.PushTokenRepository
	provider       push.Provider
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	isShuttingDown bool
}

func main() {
	cfg := config.Load()

	if cfg.LogLevel == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Str("component", "worker").Logger()

	log.Info().Msg("Starting notification worker")

	db, err := pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	queueRepo := repository.NewNotificationQueueRepository(db)
	tokenRepo := repository.NewPushTokenRepository(db)

	provider := createProvider(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	worker := &Worker{
		queueRepo: queueRepo,
		tokenRepo: tokenRepo,
		provider:  provider,
		ctx:       ctx,
		cancel:    cancel,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	worker.wg.Add(1)
	go worker.run()

	<-sigChan
	log.Info().Msg("Shutdown signal received, initiating graceful shutdown...")
	worker.shutdown()
}

func createProvider(cfg *config.Config) push.Provider {
	providers := []push.Provider{}

	if cfg.FCM.APIKey != "" {
		providers = append(providers, push.NewFCMProvider(cfg.FCM.APIKey))
		log.Info().Msg("FCM provider configured")
	}

	providers = append(providers, push.NewExpoProvider())
	log.Info().Msg("Expo provider configured")

	return push.NewMultiProvider(providers...)
}

func (w *Worker) run() {
	defer w.wg.Done()

	log.Info().Dur("interval", pollInterval).Int("batch_size", batchSize).Msg("Worker started")

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	w.processBatch()

	for {
		select {
		case <-w.ctx.Done():
			log.Info().Msg("Worker stopping")
			return
		case <-ticker.C:
			if w.isShuttingDown {
				return
			}
			w.processBatch()
		}
	}
}

func (w *Worker) processBatch() {
	items, err := w.queueRepo.Dequeue(batchSize)
	if err != nil {
		log.Error().Err(err).Msg("Failed to dequeue notifications")
		return
	}

	if len(items) == 0 {
		return
	}

	log.Info().Int("count", len(items)).Msg("Processing notification batch")

	for _, item := range items {
		if w.isShuttingDown {
			return
		}

		if err := w.processItem(item); err != nil {
			w.handleProcessingError(item, err)
		} else {
			w.handleProcessingSuccess(item)
		}
	}
}

func (w *Worker) processItem(item *domain.NotificationQueueItem) error {
	payload, err := push.ParsePayload(item.Payload)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	tokens, err := w.tokenRepo.GetActiveTokensByUserID(item.UserID)
	if err != nil {
		return fmt.Errorf("failed to get tokens: %w", err)
	}

	if len(tokens) == 0 {
		log.Warn().
			Str("user_id", item.UserID.String()).
			Str("notification_id", item.ID.String()).
			Msg("No active tokens found for user")
		return nil
	}

	ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
	defer cancel()

	var lastErr error
	invalidTokens := []string{}

	for _, tokenInfo := range tokens {
		err := w.provider.Send(ctx, tokenInfo.Token, payload.Title, payload.Body, payload.Data)
		if err == nil {
			w.tokenRepo.UpdateLastUsed(tokenInfo.Token)
			continue
		}

		if push.IsInvalidToken(err) {
			invalidTokens = append(invalidTokens, tokenInfo.Token)
			log.Warn().
				Str("token", tokenInfo.Token[:20]+"...").
				Str("reason", err.Error()).
				Msg("Invalid token detected, will be removed")
		} else {
			lastErr = err
			log.Warn().
				Err(err).
				Str("token", tokenInfo.Token[:20]+"...").
				Msg("Failed to send notification")
		}
	}

	for _, token := range invalidTokens {
		if err := w.tokenRepo.DeleteToken(token); err != nil {
			log.Error().Err(err).Str("token", token[:20]+"...").Msg("Failed to delete invalid token")
		}
	}

	if len(invalidTokens) == len(tokens) {
		return fmt.Errorf("all tokens invalid")
	}

	return lastErr
}

func (w *Worker) handleProcessingError(item *domain.NotificationQueueItem, err error) {
	log.Warn().
		Err(err).
		Str("notification_id", item.ID.String()).
		Int("retry_count", item.RetryCount).
		Msg("Failed to process notification")

	if item.RetryCount >= maxRetries-1 {
		w.moveToDeadLetter(item, err.Error())
		return
	}

	backoffDuration := calculateBackoff(item.RetryCount + 1)
	scheduledAt := time.Now().Add(backoffDuration)

	if updateErr := w.queueRepo.UpdateRetry(item.ID, err.Error(), scheduledAt); updateErr != nil {
		log.Error().
			Err(updateErr).
			Str("notification_id", item.ID.String()).
			Msg("Failed to update retry status")
	}
}

func (w *Worker) handleProcessingSuccess(item *domain.NotificationQueueItem) {
	if err := w.queueRepo.MarkSent(item.ID); err != nil {
		log.Error().
			Err(err).
			Str("notification_id", item.ID.String()).
			Msg("Failed to mark notification as sent")
	}
}

func (w *Worker) moveToDeadLetter(item *domain.NotificationQueueItem, errorMessage string) {
	log.Error().
		Str("notification_id", item.ID.String()).
		Str("user_id", item.UserID.String()).
		Str("error", errorMessage).
		Msg("Moving notification to dead letter queue")

	if err := w.tokenRepo.CreateDeadLetterEntry(item.UserID, string(item.Type), item.Payload, errorMessage); err != nil {
		log.Error().Err(err).Msg("Failed to create dead letter entry")
	}

	if err := w.queueRepo.MarkFailed(item.ID, errorMessage); err != nil {
		log.Error().Err(err).Str("notification_id", item.ID.String()).Msg("Failed to mark notification as failed")
	}
}

func (w *Worker) shutdown() {
	w.isShuttingDown = true

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	w.cancel()

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Worker shutdown completed")
	case <-shutdownCtx.Done():
		log.Warn().Msg("Worker shutdown timed out")
	}
}

func calculateBackoff(retryCount int) time.Duration {
	baseDelay := time.Second
	maxDelay := 5 * time.Minute

	delay := baseDelay * time.Duration(1<<retryCount)

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}
