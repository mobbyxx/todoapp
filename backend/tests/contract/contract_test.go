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

	"github.com/alicebob/miniredis/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/user/todo-api/internal/domain"
	"github.com/user/todo-api/internal/handler"
	"github.com/user/todo-api/internal/middleware"
	"github.com/user/todo-api/internal/service"
)

type ContractTestSuite struct {
	suite.Suite
	router        chi.Router
	userHandler   *handler.UserHandler
	todoHandler   *handler.TodoHandler
	openapiRouter routers.Router
	openAPISchema *openapi3.T
}

func TestContractSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

func (s *ContractTestSuite) SetupSuite() {
	openapi3.DefineStringFormat("email", openapi3.FormatOfStringForEmail)
	openapi3.DefineStringFormat("uuid", openapi3.FormatOfStringForUUIDOfRFC4122)

	schema := s.loadOpenAPISchema()
	s.openAPISchema = schema

	router, err := gorillamux.NewRouter(schema)
	s.Require().NoError(err)
	s.openapiRouter = router

	mockUserService := newMockUserService()
	mockTodoService := newMockTodoService()
	mockJWTService := newTestJWTService()

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
		Paths: openapi3.NewPaths(),
	}

	schema.Paths.Set("/api/v1/auth/register", &openapi3.PathItem{
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
									Type: &openapi3.Types{"object"},
									Properties: openapi3.Schemas{
										"email":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "email"}},
										"password":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, MinLength: 8}},
										"display_name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, MinLength: 2, MaxLength: openapi3.Uint64Ptr(50)}},
									},
									Required: []string{"email", "password", "display_name"},
								},
							},
						},
					},
				},
			},
			Responses: func() *openapi3.Responses {
				r := openapi3.NewResponses()
				r.Set("201", &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"object"},
										Properties: openapi3.Schemas{
											"user": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: &openapi3.Types{"object"},
													Properties: openapi3.Schemas{
														"id":           &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
														"email":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "email"}},
														"display_name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
														"created_at":   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
													},
												},
											},
											"access_token":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
											"refresh_token": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
											"expires_at":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
										},
									},
								},
							},
						},
					},
				})
				r.Set("400", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				r.Set("409", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				return r
			}(),
		},
	})

	schema.Paths.Set("/api/v1/auth/login", &openapi3.PathItem{
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
									Type: &openapi3.Types{"object"},
									Properties: openapi3.Schemas{
										"email":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "email"}},
										"password": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
									},
									Required: []string{"email", "password"},
								},
							},
						},
					},
				},
			},
			Responses: func() *openapi3.Responses {
				r := openapi3.NewResponses()
				r.Set("200", &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"object"},
										Properties: openapi3.Schemas{
											"user": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: &openapi3.Types{"object"},
													Properties: openapi3.Schemas{
														"id":           &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
														"email":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "email"}},
														"display_name": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
													},
												},
											},
											"access_token":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
											"refresh_token": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
											"expires_at":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
										},
									},
								},
							},
						},
					},
				})
				r.Set("401", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				return r
			}(),
		},
	})

	schema.Paths.Set("/api/v1/todos", &openapi3.PathItem{
		Get: &openapi3.Operation{
			OperationID: "listTodos",
			Summary:     "List all todos for the authenticated user",
			Parameters: openapi3.Parameters{
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "status",
						In:          "query",
						Description: "Filter by status",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"pending", "in_progress", "completed"}}},
					},
				},
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "page",
						In:          "query",
						Description: "Page number",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}, Default: 1}},
					},
				},
				&openapi3.ParameterRef{
					Value: &openapi3.Parameter{
						Name:        "page_size",
						In:          "query",
						Description: "Items per page",
						Schema:      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}, Default: 20, Max: openapi3.Float64Ptr(100)}},
					},
				},
			},
			Responses: func() *openapi3.Responses {
				r := openapi3.NewResponses()
				r.Set("200", &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"object"},
										Properties: openapi3.Schemas{
											"todos": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: &openapi3.Types{"array"},
													Items: &openapi3.SchemaRef{
														Value: &openapi3.Schema{
															Type: &openapi3.Types{"object"},
															Properties: openapi3.Schemas{
																"id":          &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
																"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
																"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
																"status":      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
																"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
																"created_by":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
																"version":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
																"created_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
																"updated_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
															},
														},
													},
												},
											},
											"total_count": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
											"page":        &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
											"page_size":   &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
										},
									},
								},
							},
						},
					},
				})
				r.Set("401", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				return r
			}(),
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
									Type: &openapi3.Types{"object"},
									Properties: openapi3.Schemas{
										"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, MinLength: 1, MaxLength: openapi3.Uint64Ptr(200)}},
										"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, MaxLength: openapi3.Uint64Ptr(2000)}},
										"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Enum: []interface{}{"low", "medium", "high", "urgent"}}},
										"assigned_to": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
									},
									Required: []string{"title"},
								},
							},
						},
					},
				},
			},
			Responses: func() *openapi3.Responses {
				r := openapi3.NewResponses()
				r.Set("201", &openapi3.ResponseRef{
					Value: &openapi3.Response{
						Description: new(string),
						Content: openapi3.Content{
							"application/json": &openapi3.MediaType{
								Schema: &openapi3.SchemaRef{
									Value: &openapi3.Schema{
										Type: &openapi3.Types{"object"},
										Properties: openapi3.Schemas{
											"todo": &openapi3.SchemaRef{
												Value: &openapi3.Schema{
													Type: &openapi3.Types{"object"},
													Properties: openapi3.Schemas{
														"id":          &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
														"title":       &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
														"description": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
														"status":      &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
														"priority":    &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
														"created_by":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "uuid"}},
														"version":     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}}},
														"created_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
														"updated_at":  &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}, Format: "date-time"}},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				})
				r.Set("400", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				r.Set("401", &openapi3.ResponseRef{Value: &openapi3.Response{Description: new(string)}})
				return r
			}(),
		},
	})

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

	validationCtx := openapi3.WithValidationOptions(context.Background(), openapi3.EnableSchemaFormatValidation())
	if err := openapi3filter.ValidateRequest(validationCtx, requestValidationInput); err != nil {
		return fmt.Errorf("validating request: %w", err)
	}

	return nil
}

func (s *ContractTestSuite) validateResponse(req *http.Request, resp *http.Response, body []byte) error {
	route, pathParams, err := s.openapiRouter.FindRoute(req)
	if err != nil {
		return fmt.Errorf("finding route for response: %w", err)
	}

	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status: resp.StatusCode,
		Header: resp.Header,
	}

	if len(body) > 0 {
		responseValidationInput.SetBodyBytes(body)
	}

	validationCtx := openapi3.WithValidationOptions(context.Background(), openapi3.EnableSchemaFormatValidation())
	if err := openapi3filter.ValidateResponse(validationCtx, responseValidationInput); err != nil {
		return fmt.Errorf("validating response: %w", err)
	}

	return nil
}

func (s *ContractTestSuite) TestRegisterUser_Contract() {
	reqBody := map[string]interface{}{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusCreated, rr.Code)

	err = s.validateResponse(req, rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestLoginUser_Contract() {
	registerBody := map[string]interface{}{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	registerJSON, _ := json.Marshal(registerBody)

	registerReq := httptest.NewRequest("POST", "http://localhost:8080/api/v1/auth/register", bytes.NewBuffer(registerJSON))
	registerReq.Header.Set("Content-Type", "application/json")
	registerRR := httptest.NewRecorder()
	s.router.ServeHTTP(registerRR, registerReq)
	s.Require().Equal(http.StatusCreated, registerRR.Code)

	reqBody := map[string]interface{}{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	err = s.validateResponse(req, rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestListTodos_Contract() {
	req := httptest.NewRequest("GET", "http://localhost:8080/api/v1/todos?page=1&page_size=20&status=pending", nil)
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	err := s.validateRequest(req, nil)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	err = s.validateResponse(req, rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestCreateTodo_Contract() {
	reqBody := map[string]interface{}{
		"title":       "Test Todo",
		"description": "Test Description",
		"priority":    "high",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, uuid.New().String())
	req = req.WithContext(ctx)

	err := s.validateRequest(req, jsonBody)
	s.Require().NoError(err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusCreated, rr.Code)

	err = s.validateResponse(req, rr.Result(), rr.Body.Bytes())
	s.Require().NoError(err)
}

func (s *ContractTestSuite) TestRegisterUser_InvalidEmail() {
	reqBody := map[string]interface{}{
		"email":        "invalid-email",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	err := s.validateRequest(req, jsonBody)
	s.Error(err)
}

func (s *ContractTestSuite) TestCreateTodo_ShortTitle() {
	reqBody := map[string]interface{}{
		"title": "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "http://localhost:8080/api/v1/todos", bytes.NewBuffer(jsonBody))
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

func (m *mockUserService) GetUserByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserService) UpdateLastSeen(id uuid.UUID) error {
	return nil
}

func (m *mockUserService) SoftDelete(id uuid.UUID) error {
	return nil
}

func newTestJWTService() *service.JWTService {
	mr, _ := miniredis.Run()
	rc := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return service.NewJWTService("test-secret-key-for-contract-tests", rc)
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
