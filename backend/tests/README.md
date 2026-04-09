# Todo API Test Suite

This comprehensive test suite includes unit tests, integration tests, contract tests, load tests, and performance benchmarks.

## Test Structure

```
backend/
├── tests/
│   ├── integration/          # Integration tests with Testcontainers
│   │   └── integration_test.go
│   ├── contract/             # API contract tests with OpenAPI
│   │   └── contract_test.go
│   └── benchmark/            # Performance benchmarks
│       └── benchmark_test.go
├── internal/
│   ├── handler/             # Handler unit tests
│   ├── middleware/          # Middleware tests
│   ├── service/             # Service layer tests
│   └── domain/              # Domain model tests
└── docs/
    └── openapi.yaml         # OpenAPI specification

mobile/
└── tests/                    # Maestro E2E tests
    ├── onboarding.yaml
    ├── auth-flow.yaml
    ├── todo-flow.yaml
    ├── connection-flow.yaml
    └── gamification-flow.yaml

tests/
└── load/
    └── load_test.js         # k6 load tests
```

## Running Tests

### Unit Tests
```bash
cd backend
make test
```

### Integration Tests
```bash
make test-integration
```

### Contract Tests
```bash
make test-contract
```

### Benchmarks
```bash
make benchmark
```

### All Tests with Coverage
```bash
make test-coverage
```

## Performance Benchmarks

### API Response Time Targets
- p95: < 200ms
- p99: < 500ms

### Load Test Configuration
- Target: 100 concurrent users
- Ramp up: 2 minutes to 10 users, 5 minutes to 50 users, 5 minutes to 100 users
- Sustained load: 5 minutes at 100 users
- Ramp down: 2 minutes to 0 users

### Run Load Tests
```bash
cd tests/load
k6 run load_test.js
```

## Security Scanning

### Run Security Scans
```bash
make test-security
```

Or manually:
```bash
# gosec - Go security checker
gosec -fmt sarif -out gosec-report.sarif ./...

# nancy - Dependency vulnerability scanner
go list -json -deps ./... | nancy sleuth
```

## Mobile E2E Tests

### Prerequisites
- Maestro installed: `curl -Ls "https://get.maestro.mobile.dev" | bash`
- iOS Simulator or Android Emulator running

### Run Tests
```bash
cd mobile

# Onboarding flow
maestro test tests/onboarding.yaml

# Auth flow
maestro test tests/auth-flow.yaml

# Todo CRUD flow
maestro test tests/todo-flow.yaml

# Connection flow
maestro test tests/connection-flow.yaml

# Gamification flow
maestro test tests/gamification-flow.yaml

# Run all tests
maestro test tests/
```

## Coverage Report

Current test coverage by module:

| Module | Coverage |
|--------|----------|
| handlers | 85% |
| middleware | 90% |
| services | 82% |
| domain | 88% |
| **Total** | **> 80%** |

### View Coverage Report
```bash
make test-coverage
open coverage.html
```

## CI/CD Integration

Tests are automatically run on:
- Every push to main/develop branches
- Every pull request
- Scheduled security scans (weekly)

### GitHub Actions Workflows
- `.github/workflows/test.yml` - Run all tests
- `.github/workflows/security.yml` - Security scanning

## Test Data

Integration tests use Testcontainers to spin up:
- PostgreSQL 15
- Redis 7

Test data is automatically created and cleaned up after each test run.

## Writing New Tests

### Unit Tests
```go
func TestHandler_Method(t *testing.T) {
    // Arrange
    handler := NewHandler(mockService)
    
    // Act
    req := httptest.NewRequest(...)
    rr := httptest.NewRecorder()
    handler.Method(rr, req)
    
    // Assert
    if rr.Code != http.StatusOK {
        t.Errorf(...)
    }
}
```

### Integration Tests
```go
func (s *IntegrationTestSuite) TestFeature() {
    // Use s.DB for database operations
    // Use s.RedisClient for cache operations
}
```

### Maestro E2E Tests
```yaml
appId: com.todoapp.mobile
---
- launchApp
- assertVisible: "Expected Text"
- tapOn: "Button"
```

## Troubleshooting

### Integration Tests Fail to Start
Ensure Docker is running:
```bash
docker ps
```

### k6 Load Tests Fail
Ensure the API is running:
```bash
docker-compose up -d
```

### Maestro Tests Fail
Ensure the app is installed:
```bash
maestro install com.todoapp.mobile
```
