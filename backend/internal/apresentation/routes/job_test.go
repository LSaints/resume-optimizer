package routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/internal/application/requests"
	"backend/internal/application/responses"
	"backend/internal/application/services"
	"backend/internal/apresentation/handlers"
	"backend/internal/apresentation/middleware"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func setupJobTestDB(t *testing.T) *sql.DB {
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
		);

		CREATE TABLE jobs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			raw_description TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`

	_, err = db.Exec(query)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type jobTestFixture struct {
	router         http.Handler
	db             *sql.DB
	userRepository *repositories.UserRepository
	authServices   *services.AuthServices
}

func setupJobTestFixture(t *testing.T) *jobTestFixture {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret-key-para-testes")

	db := setupJobTestDB(t)

	userRepository := repositories.NewUserRepository(db)
	userServices := services.NewUserServices(userRepository)
	userHandler := handlers.NewUserHandler(userServices)

	authServices := services.NewAuthServices(userRepository)
	authHandler := handlers.NewAuthHandler(authServices)

	authMiddleware := middleware.NewAuthMiddleware(authServices)

	jobRepository := repositories.NewJobRepository(db)
	jobServices := services.NewJobServices(jobRepository)
	jobHandler := handlers.NewJobHandler(jobServices)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

	mux.Handle(
		"POST /v1/users",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.CreateUser),
		),
	)

	mux.Handle(
		"POST /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Create),
		),
	)
	mux.Handle(
		"GET /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.List),
		),
	)
	mux.Handle(
		"GET /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.GetByID),
		),
	)
	mux.Handle(
		"PUT /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Delete),
		),
	)

	return &jobTestFixture{
		router:         mux,
		db:             db,
		userRepository: userRepository,
		authServices:   authServices,
	}
}

func insertJobUser(t *testing.T, repo *repositories.UserRepository, email, password string) entities.User {
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

func jobAuthToken(t *testing.T, fixture *jobTestFixture, email, password string) string {
	t.Helper()

	body, _ := json.Marshal(requests.LoginRequest{
		Email:    email,
		Password: password,
	})
	req := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(string(body)))
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login falhou: %d - %s", rec.Code, rec.Body.String())
	}

	var loginResp responses.LoginResponse
	err := json.NewDecoder(rec.Body).Decode(&loginResp)
	if err != nil {
		t.Fatal(err)
	}

	return loginResp.Token
}

func createJob(t *testing.T, fixture *jobTestFixture, token, title, description string) responses.JobResponse {
	t.Helper()

	reqBody, _ := json.Marshal(requests.CreateJobRequest{
		Title:          title,
		RawDescription: description,
	})

	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("criação de vaga falhou: %d - %s", rec.Code, rec.Body.String())
	}

	var resp responses.JobResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	return resp
}

func TestJob_Criar_Retorna201(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "criar@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "criar@teste.com", "senha123")

	resp := createJob(t, fixture, token, "Desenvolvedor Go", "Vaga para desenvolvedor Go senior")

	if resp.ID == uuid.Nil {
		t.Fatal("ID não gerado")
	}
	if resp.Title != "Desenvolvedor Go" {
		t.Fatalf("Title incorreto: esperado 'Desenvolvedor Go', got '%s'", resp.Title)
	}
	if resp.RawDescription != "Vaga para desenvolvedor Go senior" {
		t.Fatalf("RawDescription incorreto: esperado 'Vaga para desenvolvedor Go senior', got '%s'", resp.RawDescription)
	}
	if resp.CreatedAt.IsZero() {
		t.Fatal("CreatedAt não definido")
	}
	if resp.UpdatedAt.IsZero() {
		t.Fatal("UpdatedAt não definido")
	}
}

func TestJob_Criar_TituloVazio_Retorna400(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "titulovazio@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "titulovazio@teste.com", "senha123")

	reqBody, _ := json.Marshal(requests.CreateJobRequest{
		Title:          "",
		RawDescription: "Descricao valida",
	})

	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "título é obrigatório") {
		t.Fatalf("mensagem de erro incorreta: %s", rec.Body.String())
	}
}

func TestJob_Criar_DescricaoVazia_Retorna400(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "descricaovazia@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "descricaovazia@teste.com", "senha123")

	reqBody, _ := json.Marshal(requests.CreateJobRequest{
		Title:          "Desenvolvedor",
		RawDescription: "",
	})

	req := httptest.NewRequest("POST", "/v1/jobs", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "descrição é obrigatória") {
		t.Fatalf("mensagem de erro incorreta: %s", rec.Body.String())
	}
}

func TestJob_Listar_Retorna200(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "listar@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "listar@teste.com", "senha123")

	createJob(t, fixture, token, "Vaga 1", "Descricao 1")
	createJob(t, fixture, token, "Vaga 2", "Descricao 2")

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var jobs []responses.JobResponse
	err := json.NewDecoder(rec.Body).Decode(&jobs)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 2 {
		t.Fatalf("esperado 2 vagas, got %d", len(jobs))
	}
}

func TestJob_Listar_Vazio_RetornaArrayVazio(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "listarvazio@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "listarvazio@teste.com", "senha123")

	req := httptest.NewRequest("GET", "/v1/jobs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var jobs []responses.JobResponse
	err := json.NewDecoder(rec.Body).Decode(&jobs)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 0 {
		t.Fatalf("esperado array vazio, got %d vagas", len(jobs))
	}
}

func TestJob_Visualizar_Retorna200(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "visualizar@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "visualizar@teste.com", "senha123")

	created := createJob(t, fixture, token, "Vaga Teste", "Descricao teste")

	req := httptest.NewRequest("GET", "/v1/jobs/"+created.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.JobResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Title != "Vaga Teste" {
		t.Fatalf("Title incorreto: esperado 'Vaga Teste', got '%s'", resp.Title)
	}
	if resp.RawDescription != "Descricao teste" {
		t.Fatalf("RawDescription incorreto: esperado 'Descricao teste', got '%s'", resp.RawDescription)
	}
}

func TestJob_Atualizar_Retorna200(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "atualizar@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "atualizar@teste.com", "senha123")

	created := createJob(t, fixture, token, "Titulo Antigo", "Descricao Antiga")

	reqBody, _ := json.Marshal(requests.UpdateJobRequest{
		Title:          "Titulo Novo",
		RawDescription: "Descricao Nova",
	})

	req := httptest.NewRequest("PUT", "/v1/jobs/"+created.ID.String(), bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var updated responses.JobResponse
	err := json.NewDecoder(rec.Body).Decode(&updated)
	if err != nil {
		t.Fatal(err)
	}

	if updated.Title != "Titulo Novo" {
		t.Fatalf("Title não atualizado: esperado 'Titulo Novo', got '%s'", updated.Title)
	}
	if updated.RawDescription != "Descricao Nova" {
		t.Fatalf("RawDescription não atualizado: esperado 'Descricao Nova', got '%s'", updated.RawDescription)
	}
}

func TestJob_Excluir_Retorna204E404Depois(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "excluir@teste.com", "senha123")
	token := jobAuthToken(t, fixture, "excluir@teste.com", "senha123")

	created := createJob(t, fixture, token, "Será Excluída", "Descricao")

	reqDel := httptest.NewRequest("DELETE", "/v1/jobs/"+created.ID.String(), nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	recDel := httptest.NewRecorder()
	fixture.router.ServeHTTP(recDel, reqDel)

	if recDel.Code != http.StatusNoContent {
		t.Fatalf("esperado 204, got %d: %s", recDel.Code, recDel.Body.String())
	}

	reqGet := httptest.NewRequest("GET", "/v1/jobs/"+created.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 após excluir, got %d: %s", recGet.Code, recGet.Body.String())
	}
}

func TestJob_AcessarDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupJobTestFixture(t)
	insertJobUser(t, fixture.userRepository, "dono@teste.com", "senha123")
	tokenDono := jobAuthToken(t, fixture, "dono@teste.com", "senha123")

	created := createJob(t, fixture, tokenDono, "Vaga do Dono", "Descricao")

	insertJobUser(t, fixture.userRepository, "invasor@teste.com", "senha123")
	tokenInvasor := jobAuthToken(t, fixture, "invasor@teste.com", "senha123")

	reqGet := httptest.NewRequest("GET", "/v1/jobs/"+created.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenInvasor)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 para outro usuário, got %d: %s", recGet.Code, recGet.Body.String())
	}
}

func TestJob_SemToken_Retorna401(t *testing.T) {
	fixture := setupJobTestFixture(t)

	endpoints := []struct {
		method string
		path   string
		body   string
	}{
		{"POST", "/v1/jobs", `{"title":"Teste","rawDescription":"Teste"}`},
		{"GET", "/v1/jobs", ""},
		{"GET", "/v1/jobs/123", ""},
		{"PUT", "/v1/jobs/123", `{"title":"Teste","rawDescription":"Teste"}`},
		{"DELETE", "/v1/jobs/123", ""},
	}

	for _, ep := range endpoints {
		var req *http.Request
		if ep.body != "" {
			req = httptest.NewRequest(ep.method, ep.path, bytes.NewReader([]byte(ep.body)))
		} else {
			req = httptest.NewRequest(ep.method, ep.path, nil)
		}
		rec := httptest.NewRecorder()
		fixture.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: esperado 401, got %d", ep.method, ep.path, rec.Code)
		}
	}
}
