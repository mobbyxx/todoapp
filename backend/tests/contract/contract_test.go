package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/handler"
	"github.com/user/todo-api/internal/middleware"
)

type ContractTestSuite struct {
	suite.Suite
	router         chi.Router
	userHandler    *handler.UserHandler
	todoHandler    *handler.TodoHandler
	openapiRouter  *gorillamux.Router
	openAPISchema  *openapi3.T
}

func TestContractSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) SetupSuite() {
	schema := s.loadOpenAPISchema()
	s.openAPISchema = schema

	router, err := gorillamux.NewRouter(schema)
	s.Require().NoError(err)
	s.openapiRouter = router

	mockUserService := newMockUserService()
	mockTodoService := newMockTodoService()
	mockJWTService := newMockJWTService()

	s.userHandler = handler.NewUserHandler(mockUserService, mockJWTService)
	s.todoHandler = handler.NewTodoHandler(mockTodoService)

	s.setupRouter()
}

func (s *ContractTestSuite) loadOpenAPISchema() *openapi3.T {
	schema := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       "Todo API",
			Description: "API for Todo App with gamification features",
			Version:     "1.0.0",
		},
		Servers: openapi3.Servers{
			&openapi3.Server{
				URL:         "http://localhost:8080",
				Description: "Local development server",
			},
		},
		Paths: openapi3.Paths{},
	}

	schema.Paths["/api/v1/auth/register"] = &openapi3.PathItem{
		Post: &openapi3.Operation{
			OperationID: "registerUser",
			Summary:     "Register a new user",
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "object",
									Properties: openapi3.Schemas{
										"email":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "email"}},
										"password":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", MinLength: 8}},
										"display_name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", MinLength: 2, MaxLength: 50}},
									},
									Required: []string{"email", "password", "display_name"},
								},
							},
						},
					},
				},
			},
			Responses: openapi3.Responses{
				"201": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: "object",
										Properties: openapi3.Schemas{
											"user": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: "object",
													Properties: openapi3.Schemas{
														"id":            &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
														"email":         &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "email"}},
														"display_name":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
														"created_at":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
													},
												},
											},
											"access_token":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
											"refresh_token": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
											"expires_at":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
										},
									},
								},
							},
						},
					},
				},
				"400": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
				"409": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
			},
		},
	}

	schema.Paths["/api/v1/auth/login"] = &openapi3.PathItem{
		Post: &openapi3.Operation{
			OperationID: "loginUser",
			Summary:     "Login a user",
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "object",
									Properties: openapi3.Schemas{
										"email":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "email"}},
										"password": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
									},
									Required: []string{"email", "password"},
								},
							},
						},
					},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: "object",
										Properties: openapi3.Schemas{
											"user": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: "object",
													Properties: openapi3.Schemas{
														"id":           &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
														"email":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "email"}},
														"display_name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
													},
												},
											},
											"access_token":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
											"refresh_token": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
											"expires_at":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
										},
									},
								},
							},
						},
					},
				},
				"401": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
			},
		},
	}

	schema.Paths["/api/v1/todos"] = &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "listTodos",
			Summary:     "List all todos for the authenticated user",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "status",
						In:          "query",
						Description: "Filter by status",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"pending", "in_progress", "completed"}}},
					},
				},
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "page",
						In:          "query",
						Description: "Page number",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer", Default: 1}},
					},
				},
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "page_size",
						In:          "query",
						Description: "Items per page",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer", Default: 20, Maximum: new(float64)}},
					},
				},
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: "object",
										Properties: openapi3.Schemas{
											"todos": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: "array",
													Items: &openapi3.SchemaRef{
														Value: &openapi3.Schema{
															Type: "object",
															Properties: openapi3.Schemas{
																"id":          &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
																"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
																"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
																"status":      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
																"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
																"created_by":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
																"version":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer"}},
																"created_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
																"updated_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
															},
														},
													},
												},
											},
											"total_count": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer"}},
											"page":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer"}},
											"page_size":   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer"}},
										},
									},
								},
							},
						},
					},
				},
				"401": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
			},
		},
		Post: &openapi3.Operation{
			OperationID: "createTodo",
			Summary:     "Create a new todo",
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Required: true,
					Content: openapi3.Content{
						"application/json": &openapi3.MediaType{
							Schema: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type: "object",
									Properties: openapi3.Schemas{
										"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", MinLength: 1, MaxLength: 200}},
										"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", MaxLength: 2000}},
										"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Enum: []interface{}{"low", "medium", "high", "urgent"}}},
										"assigned_to": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
									},
									Required: []string{"title"},
								},
							},
						},
					},
				},
			},
			Responses: openapi3.Responses{
				"201": &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: "object",
										Properties: openapi3.Schemas{
											"todo": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: "object",
													Properties: openapi3.Schemas{
														"id":          &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
														"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
														"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
														"status":      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
														"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string"}},
														"created_by":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "uuid"}},
														"version":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "integer"}},
														"created_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
														"updated_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "string", Format: "date-time"}},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"400": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
				"401": &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}},
			},
		},
	}

	return schema
}

func (s *ContractTestSuite) setupRouter() {
	s.router = chi.NewRouter()
	s.router.Post("/api/v1/auth/register", s.userHandler.Register)
	s.router.Post("/api/v1/auth/login", s.userHandler.Login)
	s.router.Get("/api/v1/todos", s.todoHandler.List)
	s.router.Post("/api/v1/todos", s.todoHandler.Create)
}

func (s *ContractTestSuite) validateRequest(req *http.Request, body []byte) error {
	route, pathParams, err := s.openapiRouter.FindRoute(req)
	if err != nil {
		return fmt.Errorf("finding route: %w", err)
	}

	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	if len(body) > 0 {
		requestValidationInput.Body = body
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		return fmt.Errorf("validating request: %w", err)
	}

	return nil
}

func (s *ContractTestSuite) validateResponse(resp *http.Response, body []byte) error {
	route, pathParams, err := s.openapiRouter.FindRoute(resp.Request)
	if err != nil {
		return fmt.Errorf("finding route for response: %w", err)
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    resp.Request,
			PathParams: pathParams,
			Route:      route,
		},
		Status: resp.StatusCode,
		Header: resp.Header,
	}

	if len(body) > 0 {
		responseValidationInput.SetBodyBytes(body)
	}

	if err := openapi3filter.ValidateResponse(context.Background(), responseValidationInput); err != nil {
		return fmt.Errorf("validating response: %w", err)
	}

	return nil
}

func (s *ContractTestSuite) TestRegisterUser_Contract() {
	reqBody := map[string]interface{}{
		"email":         "test@example.com",
		"password":      "password123",
		"display_name":  "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusCreated, rr.Code)

	err = s.validateResponse(rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestLoginUser_Contract() {
	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	err = s.validateResponse(rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestListTodos_Contract() {
	req := httptest.NewRequest("GET", "/api/v1/todos?page=1&page_size=20&status=pending", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	err := s.validateRequest(req, nil)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	err = s.validateResponse(rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestCreateTodo_Contract() {
	reqBody := map[string]interface{}{
		"title":       "Test Todo",
		"description": "Test Description",
		"priority":    "high",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusCreated, rr.Code)

	err = s.validateResponse(rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestRegisterUser_InvalidEmail() {
	reqBody := map[string]interface{}{
		"email":         "invalid-email",
		"password":      "password123",
		"display_name":  "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Error(err)
}

func (s *ContractTestSuite) TestCreateTodo_ShortTitle() {
	reqBody := map[string]interface{}{
		"title": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	err := s.validateRequest(req, jsonBody)
	s.Error(err)
}

type mockUserService struct {
	users map[string]*domain.User
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserService) Register(email, password, displayName string) (*domain.User, error) {
	user := &domain.User{
		ID:          uuid.New(),
		Email:       email,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.users[user.ID.String()] = user
	return user, nil
}

func (m *mockUserService) Login(email, password string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrInvalidCredentials
}

func (m *mockUserService) GetUser(id uuid.UUID) (*domain.User, error) {
	user, ok := m.users[id.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserService) UpdateProfile(id uuid.UUID, displayName string) (*domain.User, error) {
	user, ok := m.users[id.String()]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	user.DisplayName = displayName
	user.UpdatedAt = time.Now()
	return user, nil
}

type mockJWTService struct {
	tokens map[string]bool
}

func newMockJWTService() *mockJWTService {
	return &mockJWTService{
		tokens: make(map[string]bool),
	}
}

func (m *mockJWTService) GenerateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	return &TokenPair{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (m *mockJWTService) ValidateRefreshToken(ctx context.Context, token string) (*TokenPair, error) {
	return m.GenerateTokenPair(ctx, "")
}

func (m *mockJWTService) RevokeToken(ctx context.Context, token string) error {
	return nil
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type mockTodoService struct {
	todos map[uuid.UUID]*domain.Todo
}

func newMockTodoService() *mockTodoService {
	return &mockTodoService{
		todos: make(map[uuid.UUID]*domain.Todo),
	}
}

func (m *mockTodoService) Create(userID uuid.UUID, input domain.CreateTodoInput) (*domain.Todo, error) {
	todo := &domain.Todo{
		ID:          uuid.New(),
		Title:       input.Title,
		Description: input.Description,
		Status:      domain.TodoStatusPending,
		Priority:    input.Priority,
		CreatedBy:   userID,
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if todo.Priority == "" {
		todo.Priority = domain.TodoPriorityMedium
	}
	m.todos[todo.ID] = todo
	return todo, nil
}

func (m *mockTodoService) Get(userID, todoID uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return todo, nil
}

func (m *mockTodoService) List(userID uuid.UUID, filters domain.TodoFilters) ([]*domain.Todo, int, error) {
	var result []*domain.Todo
	for _, todo := range m.todos {
		result = append(result, todo)
	}
	return result, len(result), nil
}

func (m *mockTodoService) Update(userID, todoID uuid.UUID, input domain.UpdateTodoInput, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	if input.Title != "" {
		todo.Title = input.Title
	}
	if input.Description != "" {
		todo.Description = input.Description
	}
	if input.Status != "" {
		todo.Status = input.Status
	}
	if input.Priority != "" {
		todo.Priority = input.Priority
	}
	todo.Version++
	todo.UpdatedAt = time.Now()
	return todo, nil
}

func (m *mockTodoService) Delete(userID, todoID uuid.UUID) error {
	delete(m.todos, todoID)
	return nil
}

func (m *mockTodoService) Assign(todoID, userID, assignToID uuid.UUID) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	todo.AssignedTo = &assignToID
	return todo, nil
}

func (m *mockTodoService) Complete(userID, todoID uuid.UUID, version int) (*domain.Todo, error) {
	todo, ok := m.todos[todoID]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	todo.Status = domain.TodoStatusCompleted
	todo.Version++
	return todo, nil
}
