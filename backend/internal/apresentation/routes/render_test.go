package routes

import (
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

func setupRenderTestDB(t *testing.T) *sql.DB {
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

		CREATE TABLE resumes (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			original_name TEXT NOT NULL,
			raw_text TEXT NOT NULL,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
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

		CREATE TABLE resumes_optimized (
			id TEXT PRIMARY KEY,
			resume_id TEXT NOT NULL,
			job_id TEXT NOT NULL,
			system_prompt TEXT NOT NULL,
			user_prompt TEXT NOT NULL,
			raw_text TEXT NOT NULL,
			typst_content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (resume_id) REFERENCES resumes(id),
			FOREIGN KEY (job_id) REFERENCES jobs(id)
		);
	`

	_, err = db.Exec(query)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type renderTestFixture struct {
	router         http.Handler
	db             *sql.DB
	userRepository *repositories.UserRepository
	authServices   *services.AuthServices
}

func setupRenderTestFixture(t *testing.T) *renderTestFixture {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret-key-para-testes")

	db := setupRenderTestDB(t)

	userRepository := repositories.NewUserRepository(db)
	userServices := services.NewUserServices(userRepository)
	userHandler := handlers.NewUserHandler(userServices)

	authServices := services.NewAuthServices(userRepository)
	authHandler := handlers.NewAuthHandler(authServices)

	authMiddleware := middleware.NewAuthMiddleware(authServices)

	resumeRepository := repositories.NewResumeRepository(db)
	textExtractor := services.NewTextExtractor()
	resumeServices := services.NewResumeServices(resumeRepository, textExtractor)
	resumeHandler := handlers.NewResumeHandler(resumeServices)

	jobRepository := repositories.NewJobRepository(db)
	jobServices := services.NewJobServices(jobRepository)
	jobHandler := handlers.NewJobHandler(jobServices)

	optimizationRepository := repositories.NewOptimizationRepository(db)
	geminiClient := services.NewGeminiClient()
	optimizationServices := services.NewOptimizationServices(
		optimizationRepository, resumeRepository, jobRepository, geminiClient,
	)
	optimizationHandler := handlers.NewOptimizationHandler(optimizationServices)

	typstRenderService := services.NewTypstRenderService()
	renderHandler := handlers.NewRenderHandler(optimizationServices, typstRenderService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

	mux.Handle(
		"POST /v1/users",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.CreateUser),
		),
	)
	mux.Handle(
		"POST /v1/resumes",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Create),
		),
	)
	mux.Handle(
		"GET /v1/resumes",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.List),
		),
	)
	mux.Handle(
		"POST /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Create),
		),
	)
	mux.Handle(
		"POST /v1/resumes/{resumeID}/optimize",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.Optimize),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/optimizations/{optimizationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.GetByID),
		),
	)
	mux.Handle(
		"DELETE /v1/resumes/{resumeID}/optimizations/{optimizationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.Delete),
		),
	)
	mux.Handle(
		"GET /v1/optimizations/{optimizationID}/render",
		authMiddleware.Middleware(
			http.HandlerFunc(renderHandler.RenderSVG),
		),
	)
	mux.Handle(
		"GET /v1/optimizations/{optimizationID}/render/pdf",
		authMiddleware.Middleware(
			http.HandlerFunc(renderHandler.RenderPDF),
		),
	)

	return &renderTestFixture{
		router:         mux,
		db:             db,
		userRepository: userRepository,
		authServices:   authServices,
	}
}

func insertRenderUser(t *testing.T, repo *repositories.UserRepository, email, password string) entities.User {
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

func renderAuthToken(t *testing.T, fixture *renderTestFixture, email, password string) string {
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

func createOptimization(
	t *testing.T,
	fixture *renderTestFixture,
	userID uuid.UUID,
	token string,
) (uuid.UUID, uuid.UUID) {
	t.Helper()

	resumeID := uuid.New()
	_, err := fixture.db.Exec(
		`INSERT INTO resumes (id, user_id, original_name, raw_text) VALUES (?, ?, ?, ?)`,
		resumeID.String(), userID.String(), "curriculo.pdf", "Conteudo do curriculo",
	)
	if err != nil {
		t.Fatal(err)
	}

	jobID := uuid.New()
	_, err = fixture.db.Exec(
		`INSERT INTO jobs (id, user_id, title, raw_description) VALUES (?, ?, ?, ?)`,
		jobID.String(), userID.String(), "Desenvolvedor", "Descricao da vaga",
	)
	if err != nil {
		t.Fatal(err)
	}

	optimizationID := uuid.New()
	_, err = fixture.db.Exec(
		`INSERT INTO resumes_optimized (id, resume_id, job_id, system_prompt, user_prompt, raw_text, typst_content)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		optimizationID.String(), resumeID.String(), jobID.String(),
		"prompt", "prompt", "raw", "# Hello World",
	)
	if err != nil {
		t.Fatal(err)
	}

	return optimizationID, resumeID
}

// --- Tests ---

func TestRender_SemToken_Retorna401(t *testing.T) {
	fixture := setupRenderTestFixture(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/v1/optimizations/123/render"},
		{"GET", "/v1/optimizations/123/render/pdf"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		rec := httptest.NewRecorder()
		fixture.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: esperado 401, got %d", ep.method, ep.path, rec.Code)
		}
	}
}

func TestRender_OtimizacaoInexistente_Retorna404(t *testing.T) {
	fixture := setupRenderTestFixture(t)
	insertRenderUser(t, fixture.userRepository, "inexistente@teste.com", "senha123")
	token := renderAuthToken(t, fixture, "inexistente@teste.com", "senha123")

	inexistentID := uuid.New().String()

	req := httptest.NewRequest("GET", "/v1/optimizations/"+inexistentID+"/render", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRenderPDF_OtimizacaoInexistente_Retorna404(t *testing.T) {
	fixture := setupRenderTestFixture(t)
	insertRenderUser(t, fixture.userRepository, "pdf404@teste.com", "senha123")
	token := renderAuthToken(t, fixture, "pdf404@teste.com", "senha123")

	inexistentID := uuid.New().String()

	req := httptest.NewRequest("GET", "/v1/optimizations/"+inexistentID+"/render/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRender_OtimizacaoDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupRenderTestFixture(t)

	dono := insertRenderUser(t, fixture.userRepository, "donoRender@teste.com", "senha123")
	tokenDono := renderAuthToken(t, fixture, "donoRender@teste.com", "senha123")

	optID, _ := createOptimization(t, fixture, dono.ID, tokenDono)

	insertRenderUser(t, fixture.userRepository, "invasorRender@teste.com", "senha123")
	tokenInvasor := renderAuthToken(t, fixture, "invasorRender@teste.com", "senha123")

	req := httptest.NewRequest("GET", "/v1/optimizations/"+optID.String()+"/render", nil)
	req.Header.Set("Authorization", "Bearer "+tokenInvasor)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 para outro usuário, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRenderPDF_OtimizacaoDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupRenderTestFixture(t)

	dono := insertRenderUser(t, fixture.userRepository, "donoPDF@teste.com", "senha123")
	tokenDono := renderAuthToken(t, fixture, "donoPDF@teste.com", "senha123")

	optID, _ := createOptimization(t, fixture, dono.ID, tokenDono)

	insertRenderUser(t, fixture.userRepository, "invasorPDF@teste.com", "senha123")
	tokenInvasor := renderAuthToken(t, fixture, "invasorPDF@teste.com", "senha123")

	req := httptest.NewRequest("GET", "/v1/optimizations/"+optID.String()+"/render/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+tokenInvasor)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 para outro usuário, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRender_ServiceRetorna502_QuandoTypstNaoDisponivel(t *testing.T) {
	fixture := setupRenderTestFixture(t)
	user := insertRenderUser(t, fixture.userRepository, "render502@teste.com", "senha123")
	token := renderAuthToken(t, fixture, "render502@teste.com", "senha123")

	optID, _ := createOptimization(t, fixture, user.ID, token)

	req := httptest.NewRequest("GET", "/v1/optimizations/"+optID.String()+"/render", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("esperado 502, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "renderizador typst nao disponivel") {
		t.Fatalf("mensagem de erro incorreta: %s", rec.Body.String())
	}
}

func TestRenderPDF_ServiceRetorna502_QuandoTypstNaoDisponivel(t *testing.T) {
	fixture := setupRenderTestFixture(t)
	user := insertRenderUser(t, fixture.userRepository, "pdf502@teste.com", "senha123")
	token := renderAuthToken(t, fixture, "pdf502@teste.com", "senha123")

	optID, _ := createOptimization(t, fixture, user.ID, token)

	req := httptest.NewRequest("GET", "/v1/optimizations/"+optID.String()+"/render/pdf", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("esperado 502, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "renderizador typst nao disponivel") {
		t.Fatalf("mensagem de erro incorreta: %s", rec.Body.String())
	}
}

func TestRender_ExcluirOtimizacao_Retorna404AposExclusao(t *testing.T) {
	fixture := setupRenderTestFixture(t)
	user := insertRenderUser(t, fixture.userRepository, "excluirOpt@teste.com", "senha123")
	token := renderAuthToken(t, fixture, "excluirOpt@teste.com", "senha123")

	optID, resumeID := createOptimization(t, fixture, user.ID, token)

	reqDel := httptest.NewRequest("DELETE", "/v1/resumes/"+resumeID.String()+"/optimizations/"+optID.String(), nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	recDel := httptest.NewRecorder()
	fixture.router.ServeHTTP(recDel, reqDel)

	if recDel.Code != http.StatusOK {
		t.Fatalf("esperado 200 ao excluir, got %d: %s", recDel.Code, recDel.Body.String())
	}

	reqGet := httptest.NewRequest("GET", "/v1/optimizations/"+optID.String()+"/render", nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 após excluir, got %d: %s", recGet.Code, recGet.Body.String())
	}

	reqDel2 := httptest.NewRequest("DELETE", "/v1/resumes/"+resumeID.String()+"/optimizations/"+optID.String(), nil)
	reqDel2.Header.Set("Authorization", "Bearer "+token)
	recDel2 := httptest.NewRecorder()
	fixture.router.ServeHTTP(recDel2, reqDel2)

	if recDel2.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 ao excluir novamente, got %d: %s", recDel2.Code, recDel2.Body.String())
	}
}
