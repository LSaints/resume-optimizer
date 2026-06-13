package routes

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"backend/internal/application/requests"
	"backend/internal/application/responses"
	"backend/internal/application/services"
	"backend/internal/apresentation/handlers"
	"backend/internal/apresentation/middleware"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	query := `
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL
		)
	`

	_, err = db.Exec(query)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type testFixture struct {
	router         http.Handler
	db             *sql.DB
	userRepository *repositories.UserRepository
	userServices   *services.UserServices
	authServices   *services.AuthServices
}

func setupTestFixture(t *testing.T) *testFixture {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret-key-para-testes")

	db := setupTestDB(t)

	userRepository := repositories.NewUserRepository(db)
	userServices := services.NewUserServices(userRepository)
	userHandler := handlers.NewUserHandler(userServices)

	authServices := services.NewAuthServices(userRepository)
	authHandler := handlers.NewAuthHandler(authServices)

	authMiddleware := middleware.NewAuthMiddleware(authServices)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

	mux.Handle(
		"GET /v1/users",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.GetUsers),
		),
	)
	mux.Handle(
		"GET /v1/users/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.GetUsersById),
		),
	)
	mux.Handle(
		"POST /v1/users",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.CreateUser),
		),
	)

	return &testFixture{
		router:         mux,
		db:             db,
		userRepository: userRepository,
		userServices:   userServices,
		authServices:   authServices,
	}
}

func insertUser(t *testing.T, repo *repositories.UserRepository, email, password string) entities.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	user := entities.User{
		ID:       uuid.New(),
		Name:     "Teste",
		Email:    email,
		Password: string(hash),
	}

	err = repo.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}

	return user
}

func loginRequest(t *testing.T, router http.Handler, email, password string) *httptest.ResponseRecorder {
	t.Helper()

	body, _ := json.Marshal(requests.LoginRequest{
		Email:    email,
		Password: password,
	})
	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func Test_CriarUsuario_ComSenhaHasheada(t *testing.T) {
	fixture := setupTestFixture(t)

	_, err := fixture.userServices.CreateUser(requests.CreateUserRequest{
		Name:     "Teste",
		Email:    "hash@teste.com",
		Password: "minhaSenha",
	})
	if err != nil {
		t.Fatal(err)
	}

	var senhaArmazenada string
	err = fixture.db.QueryRow(
		"SELECT password FROM users WHERE email = ?", "hash@teste.com",
	).Scan(&senhaArmazenada)
	if err != nil {
		t.Fatal(err)
	}

	if senhaArmazenada == "minhaSenha" {
		t.Fatal("senha armazenada em texto puro")
	}

	if senhaArmazenada == "" {
		t.Fatal("hash vazio")
	}

	err = bcrypt.CompareHashAndPassword([]byte(senhaArmazenada), []byte("minhaSenha"))
	if err != nil {
		t.Fatal("hash não corresponde à senha original")
	}
}

func Test_Login_ComCredenciaisValidas_RetornaToken(t *testing.T) {
	fixture := setupTestFixture(t)

	insertUser(t, fixture.userRepository, "valido@email.com", "senha123")

	rec := loginRequest(t, fixture.router, "valido@email.com", "senha123")

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var loginResp responses.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&loginResp)
	if err != nil {
		t.Fatal(err)
	}

	if loginResp.Token == "" {
		t.Fatal("token vazio")
	}

	if loginResp.ExpiresAt == "" {
		t.Fatal("expiresAt vazio")
	}
}

func Test_Login_ComSenhaErrada_Retorna401(t *testing.T) {
	fixture := setupTestFixture(t)

	insertUser(t, fixture.userRepository, "senhaerrada@email.com", "senhaCorreta")

	rec := loginRequest(t, fixture.router, "senhaerrada@email.com", "senhaErrada")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d", rec.Code)
	}
}

func Test_Login_ComEmailInexistente_Retorna401(t *testing.T) {
	fixture := setupTestFixture(t)

	rec := loginRequest(t, fixture.router, "naoexiste@email.com", "qualquer")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d", rec.Code)
	}
}

func Test_Login_ComCorpoMalformado_Retorna400(t *testing.T) {
	fixture := setupTestFixture(t)

	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader("{{invalid"))
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d", rec.Code)
	}
}

func Test_EndpointProtegido_SemToken_Retorna401(t *testing.T) {
	fixture := setupTestFixture(t)

	req := httptest.NewRequest("GET", "/v1/users", nil)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d", rec.Code)
	}
}

func Test_EndpointProtegido_ComTokenInvalido_Retorna401(t *testing.T) {
	fixture := setupTestFixture(t)

	req := httptest.NewRequest("GET", "/v1/users", nil)
	req.Header.Set("Authorization", "Bearer token-invalido-aqui")
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d", rec.Code)
	}
}

func Test_EndpointProtegido_ComTokenExpirado_Retorna401(t *testing.T) {
	fixture := setupTestFixture(t)

	claims := jwt.MapClaims{
		"userID": "123",
		"sub":    "123",
		"exp":    time.Now().Add(-1 * time.Hour).Unix(),
		"iat":    time.Now().Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key-para-testes"))
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d", rec.Code)
	}
}

func Test_EndpointProtegido_ComTokenValido_RetornaSucesso(t *testing.T) {
	fixture := setupTestFixture(t)

	insertUser(t, fixture.userRepository, "tokenvalido@email.com", "senha123")

	rec := loginRequest(t, fixture.router, "tokenvalido@email.com", "senha123")

	var loginResp responses.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&loginResp)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, req)

	if recGet.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d", recGet.Code)
	}
}
