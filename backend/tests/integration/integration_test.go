package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

type IntegrationTestSuite struct {
	suite.Suite
	postgresContainer testcontainers.Container
	redisContainer    testcontainers.Container
	DB                *pgxpool.Pool
	RedisClient       *goredis.Client
	ctx               context.Context
}

func TestIntegrationSuite(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration tests")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	s.T().Log("Starting PostgreSQL container...")
	postgresContainer, err := tcpostgres.RunContainer(s.ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithDatabase("todoapp"),
		tcpostgres.WithUsername("todoapp"),
		tcpostgres.WithPassword("todoapp"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)
	s.postgresContainer = postgresContainer

	s.T().Log("Starting Redis container...")
	redisContainer, err := tcredis.RunContainer(s.ctx,
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err)
	s.redisContainer = redisContainer

	postgresHost, err := postgresContainer.Host(s.ctx)
	s.Require().NoError(err)
	postgresPort, err := postgresContainer.MappedPort(s.ctx, "5432")
	s.Require().NoError(err)
	postgresConnStr := fmt.Sprintf("postgres://todoapp:todoapp@%s:%s/todoapp?sslmode=disable", postgresHost, postgresPort.Port())

	redisHost, err := redisContainer.Host(s.ctx)
	s.Require().NoError(err)
	redisPort, err := redisContainer.MappedPort(s.ctx, "6379")
	s.Require().NoError(err)
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort.Port())

	s.T().Log("Connecting to PostgreSQL...")
	pool, err := pgxpool.New(s.ctx, postgresConnStr)
	s.Require().NoError(err)
	s.DB = pool

	s.T().Log("Running database migrations...")
	err = s.runMigrations()
	s.Require().NoError(err)

	s.T().Log("Connecting to Redis...")
	s.RedisClient = goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})
	_, err = s.RedisClient.Ping(s.ctx).Result()
	s.Require().NoError(err)

	s.T().Log("Integration test environment ready!")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("Tearing down integration test environment...")

	if s.DB != nil {
		s.DB.Close()
	}
	if s.RedisClient != nil {
		s.RedisClient.Close()
	}
	if s.postgresContainer != nil {
		s.postgresContainer.Terminate(s.ctx)
	}
	if s.redisContainer != nil {
		s.redisContainer.Terminate(s.ctx)
	}
}

func (s *IntegrationTestSuite) runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS todos (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(200) NOT NULL,
			description TEXT,
			status VARCHAR(20) DEFAULT 'pending',
			priority VARCHAR(20) DEFAULT 'medium',
			created_by UUID REFERENCES users(id) ON DELETE CASCADE,
			assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
			due_date TIMESTAMP WITH TIME ZONE,
			version INTEGER DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS connections (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_a_id UUID REFERENCES users(id) ON DELETE CASCADE,
			user_b_id UUID REFERENCES users(id) ON DELETE CASCADE,
			status VARCHAR(20) DEFAULT 'pending',
			requested_by UUID REFERENCES users(id),
			token VARCHAR(255) UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE,
			accepted_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_a_id, user_b_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			points INTEGER DEFAULT 0,
			level INTEGER DEFAULT 1,
			streak INTEGER DEFAULT 0,
			total_todos_completed INTEGER DEFAULT 0,
			last_activity_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS points_history (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			points INTEGER NOT NULL,
			reason VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			revoked_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_created_by ON todos(created_by)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_assigned_to ON todos(assigned_to)`,
		`CREATE INDEX IF NOT EXISTS idx_todos_status ON todos(status)`,
		`CREATE INDEX IF NOT EXISTS idx_connections_user_a ON connections(user_a_id)`,
		`CREATE INDEX IF NOT EXISTS idx_connections_user_b ON connections(user_b_id)`,
		`CREATE INDEX IF NOT EXISTS idx_connections_token ON connections(token)`,
		`CREATE INDEX IF NOT EXISTS idx_points_history_user ON points_history(user_id)`,
	}

	for _, migration := range migrations {
		_, err := s.DB.Exec(s.ctx, migration)
		if err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}

func (s *IntegrationTestSuite) TestDatabaseConnection() {
	var result int
	err := s.DB.QueryRow(s.ctx, "SELECT 1").Scan(&result)
	s.Require().NoError(err)
	s.Equal(1, result)
}

func (s *IntegrationTestSuite) TestRedisConnection() {
	err := s.RedisClient.Set(s.ctx, "test_key", "test_value", time.Minute).Err()
	s.Require().NoError(err)

	val, err := s.RedisClient.Get(s.ctx, "test_key").Result()
	s.Require().NoError(err)
	s.Equal("test_value", val)
}

func (s *IntegrationTestSuite) TestUserRepository_Integration() {
	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	displayName := "Test User"
	passwordHash := "hashed_password"

	var userID string
	err := s.DB.QueryRow(s.ctx,
		`INSERT INTO users (email, password_hash, display_name) 
		 VALUES ($1, $2, $3) 
		 RETURNING id`,
		email, passwordHash, displayName,
	).Scan(&userID)
	s.Require().NoError(err)
	s.NotEmpty(userID)

	var retrievedEmail, retrievedDisplayName string
	err = s.DB.QueryRow(s.ctx,
		`SELECT email, display_name FROM users WHERE id = $1`,
		userID,
	).Scan(&retrievedEmail, &retrievedDisplayName)
	s.Require().NoError(err)
	s.Equal(email, retrievedEmail)
	s.Equal(displayName, retrievedDisplayName)
}

func (s *IntegrationTestSuite) TestTodoRepository_Integration() {
	email := fmt.Sprintf("todo_test_%d@example.com", time.Now().UnixNano())
	var userID string
	err := s.DB.QueryRow(s.ctx,
		`INSERT INTO users (email, password_hash, display_name) 
		 VALUES ($1, $2, $3) 
		 RETURNING id`,
		email, "hash", "Test User",
	).Scan(&userID)
	s.Require().NoError(err)

	title := "Test Todo"
	description := "Test Description"
	var todoID string
	err = s.DB.QueryRow(s.ctx,
		`INSERT INTO todos (title, description, created_by, status, priority) 
		 VALUES ($1, $2, $3, 'pending', 'high') 
		 RETURNING id`,
		title, description, userID,
	).Scan(&todoID)
	s.Require().NoError(err)
	s.NotEmpty(todoID)

	var retrievedTitle, retrievedStatus string
	err = s.DB.QueryRow(s.ctx,
		`SELECT title, status FROM todos WHERE id = $1`,
		todoID,
	).Scan(&retrievedTitle, &retrievedStatus)
	s.Require().NoError(err)
	s.Equal(title, retrievedTitle)
	s.Equal("pending", retrievedStatus)
}

func (s *IntegrationTestSuite) TestConnectionRepository_Integration() {
	email1 := fmt.Sprintf("conn_test1_%d@example.com", time.Now().UnixNano())
	email2 := fmt.Sprintf("conn_test2_%d@example.com", time.Now().UnixNano())

	var userID1, userID2 string
	err := s.DB.QueryRow(s.ctx,
		`INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3) RETURNING id`,
		email1, "hash", "User 1",
	).Scan(&userID1)
	s.Require().NoError(err)

	err = s.DB.QueryRow(s.ctx,
		`INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3) RETURNING id`,
		email2, "hash", "User 2",
	).Scan(&userID2)
	s.Require().NoError(err)

	token := fmt.Sprintf("token_%d", time.Now().UnixNano())
	var connID string
	err = s.DB.QueryRow(s.ctx,
		`INSERT INTO connections (user_a_id, user_b_id, requested_by, token, status, expires_at) 
		 VALUES ($1, $2, $1, $3, 'pending', NOW() + INTERVAL '7 days') 
		 RETURNING id`,
		userID1, userID2, token,
	).Scan(&connID)
	s.Require().NoError(err)
	s.NotEmpty(connID)

	var retrievedStatus string
	err = s.DB.QueryRow(s.ctx,
		`SELECT status FROM connections WHERE id = $1`,
		connID,
	).Scan(&retrievedStatus)
	s.Require().NoError(err)
	s.Equal("pending", retrievedStatus)
}

func (s *IntegrationTestSuite) TestGamificationRepository_Integration() {
	email := fmt.Sprintf("gamification_%d@example.com", time.Now().UnixNano())
	var userID string
	err := s.DB.QueryRow(s.ctx,
		`INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3) RETURNING id`,
		email, "hash", "Test User",
	).Scan(&userID)
	s.Require().NoError(err)

	_, err = s.DB.Exec(s.ctx,
		`INSERT INTO user_stats (user_id, points, level, streak) VALUES ($1, 100, 2, 5)`,
		userID,
	)
	s.Require().NoError(err)

	var points, level, streak int
	err = s.DB.QueryRow(s.ctx,
		`SELECT points, level, streak FROM user_stats WHERE user_id = $1`,
		userID,
	).Scan(&points, &level, &streak)
	s.Require().NoError(err)
	s.Equal(100, points)
	s.Equal(2, level)
	s.Equal(5, streak)

	_, err = s.DB.Exec(s.ctx,
		`INSERT INTO points_history (user_id, points, reason) VALUES ($1, 50, 'Completed todo')`,
		userID,
	)
	s.Require().NoError(err)

	var count int
	err = s.DB.QueryRow(s.ctx,
		`SELECT COUNT(*) FROM points_history WHERE user_id = $1`,
		userID,
	).Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func TestRequiresContainers(t *testing.T) {
	require.NotNil(t, &IntegrationTestSuite{}, "This test requires Docker containers")
}
