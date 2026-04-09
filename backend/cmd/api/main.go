package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/user/todo-api/config"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/handler"
	customMiddleware "github.com/user/todo-api/internal/middleware"
	"github.com/user/todo-api/internal/repository"
	"github.com/user/todo-api/internal/service"
)

func main() {
	cfg := config.Load()

	// Configure logging
	if cfg.LogLevel == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Initialize database connection
	db, err := initDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Redis connection
	redisClient, err := initRedis(cfg.Redis.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis")
	}
	defer redisClient.Close()

	// Initialize repositories
	repos := initRepositories(db)

	// Initialize services
	services := initServices(repos, redisClient, cfg)

	// Initialize handlers
	handlers := initHandlers(services)

	// Initialize rate limiter
	rateLimiter := customMiddleware.NewRateLimiter(redisClient)

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.RequestID)
	r.Use(customMiddleware.Logging)
	r.Use(customMiddleware.SecurityHeaders)
	r.Use(customMiddleware.CORS(nil))
	r.Use(rateLimiter.RateLimit)

	// Health and metrics endpoints (no auth required)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (with auth rate limiting)
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.AuthRateLimit)
			r.Mount("/auth", handlers.user.Routes())
		})

		// Public connection routes (no auth required for viewing invitations)
		r.Get("/connections/invite/{token}", handlers.connection.ValidateInvitation)

		// Protected routes (require authentication)
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Auth(services.jwt, services.apiKey))

			// User routes
			r.Mount("/users", handlers.user.ProtectedRoutes())

			// Todo routes
			r.Mount("/todos", handlers.todo.Routes())

			// Connection routes
			r.Route("/connections", func(r chi.Router) {
				r.Post("/invite", handlers.connection.CreateInvitation)
				r.Post("/invite/{token}/accept", handlers.connection.AcceptInvitation)
				r.Post("/invite/{token}/reject", handlers.connection.RejectInvitation)
				r.Get("/", handlers.connection.ListConnections)
				r.Delete("/{connectionID}", handlers.connection.RemoveConnection)
				r.Post("/qrcode/generate", handlers.connection.GenerateQRCode)
				r.Post("/qrcode/scan", handlers.connection.ScanQRCode)
			})

			// Sync routes
			r.Mount("/sync", handlers.sync.Routes())

			// Gamification routes
			r.Route("/users/me", func(r chi.Router) {
				r.Get("/stats", handlers.gamification.GetUserStats)
				r.Get("/history", handlers.gamification.GetPointsHistory)
			})

			// Reward routes
			r.Mount("/rewards", handlers.reward.Routes())

			// Goal routes
			r.Route("/goals", func(r chi.Router) {
				r.Post("/", handlers.sharedGoal.CreateGoal)
				r.Get("/", handlers.sharedGoal.ListGoals)
			})
		})
	})

	// Start server
	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	log.Info().
		Int("port", port).
		Str("domain", cfg.Server.Domain).
		Msg("Server starting")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// initDatabase initializes the PostgreSQL database connection
func initDatabase(dbURL string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().Msg("Database connection established")
	return pool, nil
}

// initRedis initializes the Redis connection
func initRedis(redisURL string) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info().Msg("Redis connection established")
	return client, nil
}

// Repositories holds all repository instances
type Repositories struct {
	user         domain.UserRepository
	todo         domain.TodoRepository
	connection   domain.ConnectionRepository
	gamification domain.GamificationRepository
	reward       domain.RewardRepository
	sharedGoal   domain.SharedGoalRepository
	sync         domain.SyncRepository
	conflict     domain.ConflictRepository
	notification domain.NotificationQueueRepository
	pushToken    *repository.PushTokenRepository
}

// initRepositories creates all repository instances
func initRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		user:         repository.NewUserRepository(db),
		todo:         repository.NewTodoRepository(db),
		connection:   repository.NewConnectionRepository(db),
		gamification: repository.NewGamificationRepository(db),
		reward:       repository.NewRewardRepository(db),
		sharedGoal:   repository.NewSharedGoalRepository(db),
		sync:         repository.NewSyncRepository(db),
		conflict:     repository.NewConflictRepository(db),
		notification: repository.NewNotificationQueueRepository(db),
		pushToken:    repository.NewPushTokenRepository(db),
	}
}

type noopAPIKeyService struct{}

func (n *noopAPIKeyService) GenerateAPIKey(userID uuid.UUID, name string, scopes []string, expiresAt *time.Time) (string, *domain.APIKey, error) {
	return "", nil, domain.ErrAPIKeyInvalid
}

func (n *noopAPIKeyService) ValidateAPIKey(key string) (*domain.APIKey, error) {
	return nil, domain.ErrAPIKeyInvalid
}

func (n *noopAPIKeyService) RevokeAPIKey(keyID uuid.UUID, reason string) error {
	return domain.ErrAPIKeyNotFound
}

func (n *noopAPIKeyService) ListAPIKeys(userID uuid.UUID) ([]*domain.APIKey, error) {
	return nil, nil
}

func (n *noopAPIKeyService) HasScope(apiKey *domain.APIKey, scope string) bool {
	return false
}

// Services holds all service instances
type Services struct {
	jwt          *service.JWTService
	user         domain.UserService
	todo         domain.TodoService
	connection   domain.ConnectionService
	gamification domain.GamificationService
	sync         domain.SyncService
	conflict     domain.ConflictService
	reward       domain.RewardService
	sharedGoal   domain.SharedGoalService
	antiCheat    domain.AntiCheatService
	notification domain.NotificationService
	apiKey       domain.APIKeyService
}

// initServices creates all service instances with proper dependency injection
func initServices(repos *Repositories, redisClient *redis.Client, cfg *config.Config) *Services {
	// JWT Service (no dependencies)
	jwtService := service.NewJWTServiceWithRotation(
		cfg.JWT.Secret,
		cfg.JWT.SecretPrevious,
		redisClient,
	)

	notificationService := service.NewNotificationService(repos.notification)

	// Anti-Cheat Service
	antiCheatConfig := domain.AntiCheatConfig{
		RateLimitWindow:      time.Minute,
		RateLimitMaxActions:  60,
		TimestampTolerance:   5 * time.Minute,
		IdempotencyTTL:       24 * time.Hour,
		MinActionGap:         time.Second,
		StatusCycleWindow:    time.Hour,
		StatusCycleThreshold: 10,
	}
	antiCheatService := service.NewAntiCheatService(redisClient, repos.todo, antiCheatConfig)

	// Gamification Service
	gamificationService := service.NewGamificationService(
		repos.gamification,
		antiCheatService,
		notificationService,
		redisClient,
	)

	// User Service
	userService := service.NewUserService(repos.user, jwtService)

	// Connection Service
	connectionService := service.NewConnectionService(
		repos.connection,
		repos.user,
		repos.notification,
		gamificationService,
		cfg.JWT.Secret,
	)

	// Todo Service
	todoService := service.NewTodoService(
		repos.todo,
		repos.connection,
		repos.user,
		gamificationService,
		notificationService,
	)

	// Conflict Service
	conflictService := service.NewConflictService(
		repos.conflict,
		repos.todo,
		repos.sync,
	)

	// Sync Service
	syncService := service.NewSyncService(
		repos.sync,
		repos.todo,
		repos.connection,
	)

	// Reward Service
	rewardService := service.NewRewardService(
		repos.reward,
		gamificationService,
	)

	// Shared Goal Service
	sharedGoalService := service.NewSharedGoalService(
		repos.sharedGoal,
		repos.connection,
		connectionService,
		gamificationService,
		notificationService,
	)

	return &Services{
		jwt:          jwtService,
		user:         userService,
		todo:         todoService,
		connection:   connectionService,
		gamification: gamificationService,
		sync:         syncService,
		conflict:     conflictService,
		reward:       rewardService,
		sharedGoal:   sharedGoalService,
		antiCheat:    antiCheatService,
		notification: notificationService,
		apiKey:       &noopAPIKeyService{},
	}
}

// Handlers holds all handler instances
type Handlers struct {
	user         *handler.UserHandler
	todo         *handler.TodoHandler
	connection   *handler.ConnectionHandler
	sync         *handler.SyncHandler
	gamification *handler.GamificationHandler
	reward       *handler.RewardHandler
	sharedGoal   *handler.SharedGoalHandler
}

// initHandlers creates all handler instances
func initHandlers(services *Services) *Handlers {
	return &Handlers{
		user:         handler.NewUserHandler(services.user, services.jwt),
		todo:         handler.NewTodoHandler(services.todo),
		connection:   handler.NewConnectionHandler(services.connection),
		sync:         handler.NewSyncHandler(services.sync, services.conflict),
		gamification: handler.NewGamificationHandler(services.gamification),
		reward:       handler.NewRewardHandler(services.reward),
		sharedGoal:   handler.NewSharedGoalHandler(services.sharedGoal, services.connection),
	}
}
