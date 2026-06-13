package services

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/application/responses"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
)

const atsScoringSystemPrompt = `Você é um especialista em recrutamento e seleção com vasto conhecimento em ATS (Applicant Tracking Systems).

Sua função é analisar a compatibilidade de um currículo com uma descrição de vaga e gerar uma avaliação objetiva.

Analise os seguintes critérios:
1. Correspondência de palavras-chave — presença de termos técnicos, ferramentas e habilidades mencionadas na vaga
2. Adequação de experiência — anos de experiência, nível hierárquico, setor de atuação
3. Compatibilidade de habilidades — hard skills e soft skills requeridas vs. apresentadas
4. Estrutura e formatação — organização, clareza, seções bem definidas
5. Resultados mensuráveis — presença de métricas, realizações quantificáveis
6. Legibilidade para ATS — uso de formato limpo, sem tabelas ou gráficos

Retorne APENAS um JSON válido no seguinte formato, sem formatação markdown:

{
  "score": 7.5,
  "summary": "Resumo da avaliação em português",
  "details": "Detalhamento com pontos fortes e oportunidades de melhoria em português"
}

A pontuação deve ser um número entre 0 e 10, com no máximo uma casa decimal.`

type atsScoreResponse struct {
	Score   float64 `json:"score"`
	Summary string  `json:"summary"`
	Details string  `json:"details"`
}

func parseEvaluationResponse(raw string) (float64, string, string, error) {
	raw = strings.TrimSpace(raw)

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var resp atsScoreResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return 0, "", "", errors.New("resposta inválida da IA")
	}

	if resp.Score < 0 || resp.Score > 10 {
		return 0, "", "", fmt.Errorf("pontuação inválida: %.1f", resp.Score)
	}

	if resp.Summary == "" || resp.Details == "" {
		return 0, "", "", errors.New("resposta inválida da IA")
	}

	return resp.Score, resp.Summary, resp.Details, nil
}

type AtsScoringServices struct {
	EvalRepo  *repositories.AtsEvaluationRepository
	ResumeRepo *repositories.ResumeRepository
	JobRepo    *repositories.JobRepository
	Gemini     *GeminiClient
}

func NewAtsScoringServices(
	evalRepo *repositories.AtsEvaluationRepository,
	resumeRepo *repositories.ResumeRepository,
	jobRepo *repositories.JobRepository,
	gemini *GeminiClient,
) *AtsScoringServices {
	return &AtsScoringServices{
		EvalRepo:   evalRepo,
		ResumeRepo: resumeRepo,
		JobRepo:    jobRepo,
		Gemini:     gemini,
	}
}

func (s *AtsScoringServices) Evaluate(userID, resumeID, jobID string) (responses.AtsEvaluationResponse, error) {
	resume, err := s.ResumeRepo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.AtsEvaluationResponse{}, errors.New("currículo não encontrado")
		}
		return responses.AtsEvaluationResponse{}, err
	}

	if resume.UserID.String() != userID {
		return responses.AtsEvaluationResponse{}, errors.New("currículo não encontrado")
	}

	job, err := s.JobRepo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.AtsEvaluationResponse{}, errors.New("vaga não encontrada")
		}
		return responses.AtsEvaluationResponse{}, err
	}

	if job.UserID.String() != userID {
		return responses.AtsEvaluationResponse{}, errors.New("vaga não encontrada")
	}

	userPrompt := "Currículo:\n" + resume.RawText + "\n\nDescrição da Vaga:\n" + job.RawDescription

	rawText, err := s.Gemini.SendPrompt(atsScoringSystemPrompt, userPrompt)
	if err != nil {
		return responses.AtsEvaluationResponse{}, err
	}

	score, summary, details, err := parseEvaluationResponse(rawText)
	if err != nil {
		return responses.AtsEvaluationResponse{}, err
	}

	eval := entities.AtsEvaluation{
		ID:          uuid.New(),
		ResumeID:    resume.ID,
		JobID:       job.ID,
		Score:       score,
		Summary:     summary,
		Details:     details,
		RawResponse: rawText,
		CreatedAt:   time.Now(),
	}

	err = s.EvalRepo.Create(eval)
	if err != nil {
		return responses.AtsEvaluationResponse{}, err
	}

	return s.toResponse(eval), nil
}

func (s *AtsScoringServices) ListByResume(userID, resumeID string) ([]responses.AtsEvaluationSummaryResponse, error) {
	resume, err := s.ResumeRepo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("currículo não encontrado")
		}
		return nil, err
	}

	if resume.UserID.String() != userID {
		return nil, errors.New("currículo não encontrado")
	}

	evals, err := s.EvalRepo.GetByResumeID(resumeID)
	if err != nil {
		return nil, err
	}

	result := make([]responses.AtsEvaluationSummaryResponse, 0, len(evals))
	for _, eval := range evals {
		result = append(result, responses.AtsEvaluationSummaryResponse{
			ID:        eval.ID.String(),
			ResumeID:  eval.ResumeID.String(),
			JobID:     eval.JobID.String(),
			Score:     eval.Score,
			Summary:   eval.Summary,
			CreatedAt: eval.CreatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *AtsScoringServices) GetByID(userID, evaluationID string) (responses.AtsEvaluationResponse, error) {
	eval, err := s.EvalRepo.GetByID(evaluationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
		}
		return responses.AtsEvaluationResponse{}, err
	}

	resume, err := s.ResumeRepo.GetByID(eval.ResumeID.String())
	if err != nil {
		return responses.AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
	}

	if resume.UserID.String() != userID {
		return responses.AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
	}

	return s.toResponse(eval), nil
}

func (s *AtsScoringServices) toResponse(eval entities.AtsEvaluation) responses.AtsEvaluationResponse {
	return responses.AtsEvaluationResponse{
		ID:        eval.ID.String(),
		ResumeID:  eval.ResumeID.String(),
		JobID:     eval.JobID.String(),
		Score:     eval.Score,
		Summary:   eval.Summary,
		Details:   eval.Details,
		CreatedAt: eval.CreatedAt.Format(time.RFC3339),
	}
}
