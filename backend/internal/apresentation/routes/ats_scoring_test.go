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
	"time"

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

func setupAtsTestDB(t *testing.T) *sql.DB {
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

		CREATE TABLE ats_evaluations (
			id TEXT PRIMARY KEY,
			resume_id TEXT NOT NULL,
			job_id TEXT NOT NULL,
			score REAL NOT NULL,
			summary TEXT NOT NULL,
			details TEXT NOT NULL,
			raw_response TEXT NOT NULL,
			breakdown_keyword_match REAL DEFAULT 0,
			breakdown_technical REAL DEFAULT 0,
			breakdown_experience REAL DEFAULT 0,
			breakdown_impact REAL DEFAULT 0,
			breakdown_readability REAL DEFAULT 0,
			matched_keywords TEXT DEFAULT '[]',
			missing_keywords TEXT DEFAULT '[]',
			recommendations TEXT DEFAULT '[]',
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

type atsTestFixture struct {
	router         http.Handler
	db             *sql.DB
	userRepository *repositories.UserRepository
	authServices   *services.AuthServices
}

func setupAtsTestFixture(t *testing.T) *atsTestFixture {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret-key-para-testes")

	db := setupAtsTestDB(t)

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

	evalRepository := repositories.NewAtsEvaluationRepository(db)
	geminiClient := services.NewGeminiClient()
	atsScoringServices := services.NewAtsScoringServices(evalRepository, resumeRepository, jobRepository, geminiClient)
	atsScoringHandler := handlers.NewAtsScoringHandler(atsScoringServices)

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
		"POST /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Create),
		),
	)

	mux.Handle(
		"POST /v1/resumes/{resumeID}/evaluate",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.Evaluate),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/evaluations",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.ListByResume),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/evaluations/{evaluationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.GetByID),
		),
	)

	return &atsTestFixture{
		router:         mux,
		db:             db,
		userRepository: userRepository,
		authServices:   authServices,
	}
}

func insertAtsUser(t *testing.T, repo *repositories.UserRepository, email, password string) entities.User {
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

func atsAuthToken(t *testing.T, fixture *atsTestFixture, email, password string) string {
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

func insertDirectResume(t *testing.T, db *sql.DB, userID uuid.UUID, rawText string) string {
	t.Helper()

	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO resumes (id, user_id, original_name, raw_text) VALUES (?, ?, ?, ?)`,
		id, userID.String(), "curriculo.pdf", rawText,
	)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func insertDirectJob(t *testing.T, db *sql.DB, userID uuid.UUID, title, description string) string {
	t.Helper()

	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO jobs (id, user_id, title, raw_description) VALUES (?, ?, ?, ?)`,
		id, userID.String(), title, description,
	)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func insertDirectEvaluation(t *testing.T, db *sql.DB, resumeID, jobID string) string {
	t.Helper()

	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO ats_evaluations (id, resume_id, job_id, score, summary, details, raw_response, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, resumeID, jobID, 7.5, "Resumo da avaliação", "Detalhamento completo", `{"score":7.5}`,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func insertDirectEvaluationComBreakdown(t *testing.T, db *sql.DB, resumeID, jobID string) string {
	t.Helper()

	id := uuid.New().String()
	_, err := db.Exec(
		`INSERT INTO ats_evaluations
		 (id, resume_id, job_id, score, summary, details, raw_response,
		  breakdown_keyword_match, breakdown_technical, breakdown_experience,
		  breakdown_impact, breakdown_readability,
		  matched_keywords, missing_keywords, recommendations, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, resumeID, jobID,
		8.3, "Resumo da avaliação", "Detalhamento completo", `{"score":8.3}`,
		2.6, 2.1, 1.6, 1.2, 0.8,
		`["Go","SQL"]`, `["Python"]`, `["Adicione experiência em Python"]`,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

// --- Evaluate: error cases ---

func TestAtsEvaluation_Evaluate_SemToken_Retorna401(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	req := httptest.NewRequest("POST", "/v1/resumes/123/evaluate", nil)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_SemJobId_Retorna400(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	insertAtsUser(t, fixture.userRepository, "semjobid@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "semjobid@teste.com", "senha123")

	req := httptest.NewRequest("POST", "/v1/resumes/123/evaluate", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "jobId é obrigatório") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_JSONInvalido_Retorna400(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	insertAtsUser(t, fixture.userRepository, "jsoninv@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "jsoninv@teste.com", "senha123")

	req := httptest.NewRequest("POST", "/v1/resumes/123/evaluate", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "json inválido") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_ResumeInexistente_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	insertAtsUser(t, fixture.userRepository, "res404@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "res404@teste.com", "senha123")

	body, _ := json.Marshal(requests.EvaluateResumeRequest{JobID: uuid.New().String()})
	req := httptest.NewRequest("POST", "/v1/resumes/"+uuid.New().String()+"/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "currículo não encontrado") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_JobInexistente_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "job404@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "job404@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo do curriculo")

	body, _ := json.Marshal(requests.EvaluateResumeRequest{JobID: uuid.New().String()})
	req := httptest.NewRequest("POST", "/v1/resumes/"+resumeID+"/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vaga não encontrada") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_ResumeDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	dono := insertAtsUser(t, fixture.userRepository, "donoAts@teste.com", "senha123")
	resumeID := insertDirectResume(t, fixture.db, dono.ID, "Curriculo do dono")

	insertAtsUser(t, fixture.userRepository, "invasorAts@teste.com", "senha123")
	tokenInvasor := atsAuthToken(t, fixture, "invasorAts@teste.com", "senha123")

	body, _ := json.Marshal(requests.EvaluateResumeRequest{JobID: uuid.New().String()})
	req := httptest.NewRequest("POST", "/v1/resumes/"+resumeID+"/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenInvasor)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "currículo não encontrado") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_JobDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	dono := insertAtsUser(t, fixture.userRepository, "donoJob@teste.com", "senha123")
	jobID := insertDirectJob(t, fixture.db, dono.ID, "Vaga do dono", "Descricao")

	user := insertAtsUser(t, fixture.userRepository, "invasorJob@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "invasorJob@teste.com", "senha123")
	resumeID := insertDirectResume(t, fixture.db, user.ID, "Meu curriculo")

	body, _ := json.Marshal(requests.EvaluateResumeRequest{JobID: jobID})
	req := httptest.NewRequest("POST", "/v1/resumes/"+resumeID+"/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "vaga não encontrada") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Evaluate_SemGeminiKey_Retorna500(t *testing.T) {
	os.Unsetenv("GEMINI_API_KEY")

	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "semkey@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "semkey@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo do curriculo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Desenvolvedor", "Descricao da vaga")

	body, _ := json.Marshal(requests.EvaluateResumeRequest{JobID: jobID})
	req := httptest.NewRequest("POST", "/v1/resumes/"+resumeID+"/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("esperado 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "serviço de IA não configurado") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

// --- List / Get: success cases with direct DB insertion ---

func TestAtsEvaluation_Listar_Retorna200(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "listarats@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "listarats@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Vaga", "Descricao")
	evalID := insertDirectEvaluation(t, fixture.db, resumeID, jobID)

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var evals []responses.AtsEvaluationSummaryResponse
	err := json.NewDecoder(rec.Body).Decode(&evals)
	if err != nil {
		t.Fatal(err)
	}

	if len(evals) != 1 {
		t.Fatalf("esperado 1 avaliação, got %d", len(evals))
	}

	if evals[0].ID != evalID {
		t.Fatalf("ID incorreto: esperado %s, got %s", evalID, evals[0].ID)
	}
	if evals[0].ResumeID != resumeID {
		t.Fatalf("ResumeID incorreto: esperado %s, got %s", resumeID, evals[0].ResumeID)
	}
	if evals[0].JobID != jobID {
		t.Fatalf("JobID incorreto: esperado %s, got %s", jobID, evals[0].JobID)
	}
	if evals[0].Score != 7.5 {
		t.Fatalf("Score incorreto: esperado 7.5, got %.1f", evals[0].Score)
	}
	if evals[0].Summary != "Resumo da avaliação" {
		t.Fatalf("Summary incorreto: esperado 'Resumo da avaliação', got '%s'", evals[0].Summary)
	}
	jsonBytes, _ := json.Marshal(evals[0])
	if strings.Contains(string(jsonBytes), "details") {
		t.Fatal("listagem não deve conter campo details")
	}
}

func TestAtsEvaluation_Listar_Vazio_RetornaArrayVazio(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "vazio@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "vazio@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var evals []responses.AtsEvaluationSummaryResponse
	err := json.NewDecoder(rec.Body).Decode(&evals)
	if err != nil {
		t.Fatal(err)
	}

	if len(evals) != 0 {
		t.Fatalf("esperado array vazio, got %d avaliações", len(evals))
	}
}

func TestAtsEvaluation_Visualizar_Retorna200(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "visualizarats@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "visualizarats@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Vaga", "Descricao")
	evalID := insertDirectEvaluation(t, fixture.db, resumeID, jobID)

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations/"+evalID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.AtsEvaluationResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.ID != evalID {
		t.Fatalf("ID incorreto: esperado %s, got %s", evalID, resp.ID)
	}
	if resp.ResumeID != resumeID {
		t.Fatalf("ResumeID incorreto: esperado %s, got %s", resumeID, resp.ResumeID)
	}
	if resp.JobID != jobID {
		t.Fatalf("JobID incorreto: esperado %s, got %s", jobID, resp.JobID)
	}
	if resp.Score != 7.5 {
		t.Fatalf("Score incorreto: esperado 7.5, got %.1f", resp.Score)
	}
	if resp.Summary != "Resumo da avaliação" {
		t.Fatalf("Summary incorreto: esperado 'Resumo da avaliação', got '%s'", resp.Summary)
	}
	if resp.Details != "Detalhamento completo" {
		t.Fatalf("Details incorreto: esperado 'Detalhamento completo', got '%s'", resp.Details)
	}
	if resp.CreatedAt == "" {
		t.Fatal("CreatedAt não definido")
	}
}

// --- List / Get: error cases ---

func TestAtsEvaluation_Listar_DeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	dono := insertAtsUser(t, fixture.userRepository, "donoList@teste.com", "senha123")
	resumeID := insertDirectResume(t, fixture.db, dono.ID, "Curriculo")

	insertAtsUser(t, fixture.userRepository, "invasorList@teste.com", "senha123")
	tokenInvasor := atsAuthToken(t, fixture, "invasorList@teste.com", "senha123")

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations", nil)
	req.Header.Set("Authorization", "Bearer "+tokenInvasor)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "currículo não encontrado") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Visualizar_DeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	dono := insertAtsUser(t, fixture.userRepository, "donoVis@teste.com", "senha123")
	resumeID := insertDirectResume(t, fixture.db, dono.ID, "Curriculo")
	jobID := insertDirectJob(t, fixture.db, dono.ID, "Vaga", "Desc")
	evalID := insertDirectEvaluation(t, fixture.db, resumeID, jobID)

	insertAtsUser(t, fixture.userRepository, "invasorVis@teste.com", "senha123")
	tokenInvasor := atsAuthToken(t, fixture, "invasorVis@teste.com", "senha123")

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations/"+evalID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenInvasor)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "avaliação não encontrada") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

func TestAtsEvaluation_Visualizar_Inexistente_Retorna404(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "inexistente@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "inexistente@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations/"+uuid.New().String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "avaliação não encontrada") {
		t.Fatalf("mensagem incorreta: %s", rec.Body.String())
	}
}

// --- Sem Token ---

func TestAtsEvaluation_Listar_SemToken_Retorna401(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	req := httptest.NewRequest("GET", "/v1/resumes/123/evaluations", nil)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAtsEvaluation_Visualizar_SemToken_Retorna401(t *testing.T) {
	fixture := setupAtsTestFixture(t)

	req := httptest.NewRequest("GET", "/v1/resumes/123/evaluations/456", nil)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Enriched Evaluation ---

func TestAtsEvaluation_Visualizar_ComBreakdown_Retorna200(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "breakdown@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "breakdown@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Vaga", "Descricao")
	evalID := insertDirectEvaluationComBreakdown(t, fixture.db, resumeID, jobID)

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations/"+evalID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.AtsEvaluationResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.ID != evalID {
		t.Fatalf("ID incorreto: esperado %s, got %s", evalID, resp.ID)
	}

	if resp.BreakdownKeywordMatch != 2.6 {
		t.Fatalf("BreakdownKeywordMatch: esperado 2.6, got %.1f", resp.BreakdownKeywordMatch)
	}
	if resp.BreakdownTechnical != 2.1 {
		t.Fatalf("BreakdownTechnical: esperado 2.1, got %.1f", resp.BreakdownTechnical)
	}
	if resp.BreakdownExperience != 1.6 {
		t.Fatalf("BreakdownExperience: esperado 1.6, got %.1f", resp.BreakdownExperience)
	}
	if resp.BreakdownImpact != 1.2 {
		t.Fatalf("BreakdownImpact: esperado 1.2, got %.1f", resp.BreakdownImpact)
	}
	if resp.BreakdownReadability != 0.8 {
		t.Fatalf("BreakdownReadability: esperado 0.8, got %.1f", resp.BreakdownReadability)
	}

	if len(resp.MatchedKeywords) != 2 || resp.MatchedKeywords[0] != "Go" {
		t.Fatalf("matchedKeywords incorreto: %v", resp.MatchedKeywords)
	}
	if len(resp.MissingKeywords) != 1 || resp.MissingKeywords[0] != "Python" {
		t.Fatalf("missingKeywords incorreto: %v", resp.MissingKeywords)
	}
	if len(resp.Recommendations) != 1 || resp.Recommendations[0] != "Adicione experiência em Python" {
		t.Fatalf("recommendations incorreto: %v", resp.Recommendations)
	}

	soma := resp.BreakdownKeywordMatch + resp.BreakdownTechnical + resp.BreakdownExperience + resp.BreakdownImpact + resp.BreakdownReadability
	if soma-resp.Score > 0.1 || resp.Score-soma > 0.1 {
		t.Fatalf("soma dos breakdowns (%.1f) difere do score (%.1f)", soma, resp.Score)
	}
}

func TestAtsEvaluation_Visualizar_AvaliacaoAntiga_SemBreakdown_Retorna200(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "antiga@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "antiga@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Vaga", "Descricao")
	evalID := insertDirectEvaluation(t, fixture.db, resumeID, jobID)

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations/"+evalID, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.AtsEvaluationResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.ID != evalID {
		t.Fatalf("ID incorreto: esperado %s, got %s", evalID, resp.ID)
	}

	if resp.BreakdownKeywordMatch != 0 {
		t.Fatalf("BreakdownKeywordMatch: esperado 0, got %.1f", resp.BreakdownKeywordMatch)
	}
	if resp.BreakdownTechnical != 0 {
		t.Fatalf("BreakdownTechnical: esperado 0, got %.1f", resp.BreakdownTechnical)
	}
	if resp.BreakdownExperience != 0 {
		t.Fatalf("BreakdownExperience: esperado 0, got %.1f", resp.BreakdownExperience)
	}
	if resp.BreakdownImpact != 0 {
		t.Fatalf("BreakdownImpact: esperado 0, got %.1f", resp.BreakdownImpact)
	}
	if resp.BreakdownReadability != 0 {
		t.Fatalf("BreakdownReadability: esperado 0, got %.1f", resp.BreakdownReadability)
	}

	if len(resp.MatchedKeywords) != 0 {
		t.Fatalf("matchedKeywords: esperado vazio, got %v", resp.MatchedKeywords)
	}
	if len(resp.MissingKeywords) != 0 {
		t.Fatalf("missingKeywords: esperado vazio, got %v", resp.MissingKeywords)
	}
	if len(resp.Recommendations) != 0 {
		t.Fatalf("recommendations: esperado vazio, got %v", resp.Recommendations)
	}
}

func TestAtsEvaluation_Listar_NaoContemBreakdownNemListas(t *testing.T) {
	fixture := setupAtsTestFixture(t)
	user := insertAtsUser(t, fixture.userRepository, "listbreak@teste.com", "senha123")
	token := atsAuthToken(t, fixture, "listbreak@teste.com", "senha123")

	resumeID := insertDirectResume(t, fixture.db, user.ID, "Conteudo")
	jobID := insertDirectJob(t, fixture.db, user.ID, "Vaga", "Descricao")
	insertDirectEvaluationComBreakdown(t, fixture.db, resumeID, jobID)

	req := httptest.NewRequest("GET", "/v1/resumes/"+resumeID+"/evaluations", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var evals []responses.AtsEvaluationSummaryResponse
	err := json.NewDecoder(rec.Body).Decode(&evals)
	if err != nil {
		t.Fatal(err)
	}

	if len(evals) != 1 {
		t.Fatalf("esperado 1 avaliação, got %d", len(evals))
	}

	jsonBytes, _ := json.Marshal(evals[0])
	bodyStr := string(jsonBytes)

	camposProibidos := []string{"breakdownKeywordMatch", "breakdownTechnical", "breakdownExperience", "breakdownImpact", "breakdownReadability", "matchedKeywords", "missingKeywords", "recommendations", "details"}
	for _, campo := range camposProibidos {
		if strings.Contains(bodyStr, campo) {
			t.Fatalf("listagem não deve conter campo '%s': %s", campo, bodyStr)
		}
	}
}
