package routes

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

func setupResumeTestDB(t *testing.T) *sql.DB {
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
	`

	_, err = db.Exec(query)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

type resumeTestFixture struct {
	router         http.Handler
	db             *sql.DB
	userRepository *repositories.UserRepository
	authServices   *services.AuthServices
}

func setupResumeTestFixture(t *testing.T) *resumeTestFixture {
	t.Helper()

	os.Setenv("JWT_SECRET", "test-secret-key-para-testes")

	db := setupResumeTestDB(t)

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
		"GET /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.GetByID),
		),
	)
	mux.Handle(
		"PUT /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Delete),
		),
	)

	return &resumeTestFixture{
		router:         mux,
		db:             db,
		userRepository: userRepository,
		authServices:   authServices,
	}
}

func insertResumeUser(t *testing.T, repo *repositories.UserRepository, email, password string) entities.User {
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

func resumeAuthToken(t *testing.T, fixture *resumeTestFixture, email, password string) string {
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

func createMultipartBody(t *testing.T, fileName string, fileData []byte) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}

	_, err = part.Write(fileData)
	if err != nil {
		t.Fatal(err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes(), writer.FormDataContentType()
}

func createTestPDF(t *testing.T, text string) []byte {
	t.Helper()

	var buf bytes.Buffer

	buf.WriteString("%PDF-1.4\n")

	obj1 := buf.Len()
	buf.WriteString("1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n")

	obj2 := buf.Len()
	buf.WriteString("2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n")

	obj3 := buf.Len()
	buf.WriteString("3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R/Resources<</Font<</F1 5 0 R>>>>>>endobj\n")

	content := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", text)
	streamLen := len(content)
	obj4 := buf.Len()
	fmt.Fprintf(&buf, "4 0 obj<</Length %d>>stream\n%s\nendstream\nendobj\n", streamLen, content)

	obj5 := buf.Len()
	buf.WriteString("5 0 obj<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>endobj\n")

	xrefOffset := buf.Len()

	fmt.Fprintf(&buf, "xref\n0 6\n0000000000 65535 f \n")
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj1)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj2)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj3)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj4)
	fmt.Fprintf(&buf, "%010d 00000 n \n", obj5)

	fmt.Fprintf(&buf, "trailer<</Size 6/Root 1 0 R>>\n")
	fmt.Fprintf(&buf, "startxref\n%d\n%%%%EOF\n", xrefOffset)

	return buf.Bytes()
}

func createTestDOCX(t *testing.T, text string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	part, err := zipWriter.Create("[Content_Types].xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(contentTypes))
	if err != nil {
		t.Fatal(err)
	}

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	part, err = zipWriter.Create("_rels/.rels")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(rels))
	if err != nil {
		t.Fatal(err)
	}

	docRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

	part, err = zipWriter.Create("word/_rels/document.xml.rels")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(docRels))
	if err != nil {
		t.Fatal(err)
	}

	document := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r>
        <w:t>%s</w:t>
      </w:r>
    </w:p>
  </w:body>
</w:document>`, text)

	part, err = zipWriter.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte(document))
	if err != nil {
		t.Fatal(err)
	}

	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

// --- Tests ---

func TestResume_UploadPDF_Retorna200(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	user := insertResumeUser(t, fixture.userRepository, "pdf@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "pdf@teste.com", "senha123")

	pdfData := createTestPDF(t, "Conteudo do PDF")
	body, contentType := createMultipartBody(t, "curriculo.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.ResumeResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if resp.ID == uuid.Nil {
		t.Fatal("ID não gerado")
	}
	if resp.UserID != user.ID {
		t.Fatalf("UserID incorreto: esperado %s, got %s", user.ID, resp.UserID)
	}
	if resp.OriginalName != "curriculo.pdf" {
		t.Fatalf("OriginalName incorreto: esperado curriculo.pdf, got %s", resp.OriginalName)
	}
	if !strings.Contains(resp.RawText, "Conteudo do PDF") {
		t.Fatalf("RawText não contém texto esperado: %s", resp.RawText)
	}
	if resp.UploadedAt.IsZero() {
		t.Fatal("UploadedAt não definido")
	}
}

func TestResume_UploadDOCX_Retorna200(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "docx@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "docx@teste.com", "senha123")

	docxData := createTestDOCX(t, "Conteudo do DOCX")
	body, contentType := createMultipartBody(t, "curriculo.docx", docxData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp responses.ResumeResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(resp.RawText, "Conteudo do DOCX") {
		t.Fatalf("RawText não contém texto esperado: %s", resp.RawText)
	}
}

func TestResume_Listar_SemRawText(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "listar@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "listar@teste.com", "senha123")

	pdfData := createTestPDF(t, "Listagem")
	body, contentType := createMultipartBody(t, "cv.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	reqList := httptest.NewRequest("GET", "/v1/resumes", nil)
	reqList.Header.Set("Authorization", "Bearer "+token)
	recList := httptest.NewRecorder()
	fixture.router.ServeHTTP(recList, reqList)

	if recList.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", recList.Code, recList.Body.String())
	}

	var resumes []responses.ResumeSummaryResponse
	err := json.NewDecoder(recList.Body).Decode(&resumes)
	if err != nil {
		t.Fatal(err)
	}

	if len(resumes) != 1 {
		t.Fatalf("esperado 1 currículo, got %d", len(resumes))
	}

	jsonBytes, _ := json.Marshal(resumes[0])
	if strings.Contains(string(jsonBytes), "rawText") {
		t.Fatal("listagem não deve conter rawText")
	}
}

func TestResume_Visualizar_ComRawText(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "visualizar@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "visualizar@teste.com", "senha123")

	pdfData := createTestPDF(t, "Visualizacao Individual")
	body, contentType := createMultipartBody(t, "cv.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created responses.ResumeResponse
	json.NewDecoder(rec.Body).Decode(&created)

	reqGet := httptest.NewRequest("GET", "/v1/resumes/"+created.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", recGet.Code, recGet.Body.String())
	}

	var resp responses.ResumeResponse
	err := json.NewDecoder(recGet.Body).Decode(&resp)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(resp.RawText, "Visualizacao Individual") {
		t.Fatalf("RawText não contém texto esperado: %s", resp.RawText)
	}
}

func TestResume_Atualizar_SubstituiTexto(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "atualizar@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "atualizar@teste.com", "senha123")

	pdfData := createTestPDF(t, "Texto Antigo")
	body, contentType := createMultipartBody(t, "cv.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created responses.ResumeResponse
	json.NewDecoder(rec.Body).Decode(&created)

	pdfData2 := createTestPDF(t, "Texto Novo")
	body2, contentType2 := createMultipartBody(t, "cv_atualizado.pdf", pdfData2)

	reqPut := httptest.NewRequest("PUT", "/v1/resumes/"+created.ID.String(), bytes.NewReader(body2))
	reqPut.Header.Set("Content-Type", contentType2)
	reqPut.Header.Set("Authorization", "Bearer "+token)
	recPut := httptest.NewRecorder()
	fixture.router.ServeHTTP(recPut, reqPut)

	if recPut.Code != http.StatusOK {
		t.Fatalf("esperado 200, got %d: %s", recPut.Code, recPut.Body.String())
	}

	var updated responses.ResumeResponse
	err := json.NewDecoder(recPut.Body).Decode(&updated)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(updated.RawText, "Texto Novo") {
		t.Fatalf("RawText não foi atualizado: %s", updated.RawText)
	}
	if strings.Contains(updated.RawText, "Texto Antigo") {
		t.Fatal("RawText ainda contém texto antigo")
	}
	if updated.OriginalName != "cv_atualizado.pdf" {
		t.Fatalf("OriginalName não atualizado: %s", updated.OriginalName)
	}
}

func TestResume_Excluir_RemoveERetorna404(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "excluir@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "excluir@teste.com", "senha123")

	pdfData := createTestPDF(t, "Será Excluído")
	body, contentType := createMultipartBody(t, "cv.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created responses.ResumeResponse
	json.NewDecoder(rec.Body).Decode(&created)

	reqDel := httptest.NewRequest("DELETE", "/v1/resumes/"+created.ID.String(), nil)
	reqDel.Header.Set("Authorization", "Bearer "+token)
	recDel := httptest.NewRecorder()
	fixture.router.ServeHTTP(recDel, reqDel)

	if recDel.Code != http.StatusNoContent {
		t.Fatalf("esperado 204, got %d: %s", recDel.Code, recDel.Body.String())
	}

	reqGet := httptest.NewRequest("GET", "/v1/resumes/"+created.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+token)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 após excluir, got %d: %s", recGet.Code, recGet.Body.String())
	}
}

func TestResume_AcessarDeOutroUsuario_Retorna404(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "dono@teste.com", "senha123")
	tokenDono := resumeAuthToken(t, fixture, "dono@teste.com", "senha123")

	pdfData := createTestPDF(t, "Currículo do Dono")
	body, contentType := createMultipartBody(t, "cv.pdf", pdfData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+tokenDono)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("esperado 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created responses.ResumeResponse
	json.NewDecoder(rec.Body).Decode(&created)

	insertResumeUser(t, fixture.userRepository, "invasor@teste.com", "senha123")
	tokenInvasor := resumeAuthToken(t, fixture, "invasor@teste.com", "senha123")

	reqGet := httptest.NewRequest("GET", "/v1/resumes/"+created.ID.String(), nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenInvasor)
	recGet := httptest.NewRecorder()
	fixture.router.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusNotFound {
		t.Fatalf("esperado 404 para outro usuário, got %d: %s", recGet.Code, recGet.Body.String())
	}
}

func TestResume_UploadPNG_Retorna400(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "png@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "png@teste.com", "senha123")

	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	body, contentType := createMultipartBody(t, "imagem.png", pngData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResume_UploadMaiorQue10MB_Retorna400(t *testing.T) {
	fixture := setupResumeTestFixture(t)
	insertResumeUser(t, fixture.userRepository, "grande@teste.com", "senha123")
	token := resumeAuthToken(t, fixture, "grande@teste.com", "senha123")

	largeData := make([]byte, 11<<20)
	body, contentType := createMultipartBody(t, "grande.pdf", largeData)

	req := httptest.NewRequest("POST", "/v1/resumes", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	fixture.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResume_SemToken_Retorna401(t *testing.T) {
	fixture := setupResumeTestFixture(t)

	endpoints := []struct {
		method string
		path   string
		body   io.Reader
	}{
		{"POST", "/v1/resumes", nil},
		{"GET", "/v1/resumes", nil},
		{"GET", "/v1/resumes/123", nil},
		{"PUT", "/v1/resumes/123", nil},
		{"DELETE", "/v1/resumes/123", nil},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, ep.body)
		rec := httptest.NewRecorder()
		fixture.router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: esperado 401, got %d", ep.method, ep.path, rec.Code)
		}
	}
}
