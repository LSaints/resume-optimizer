package services

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"backend/internal/application/responses"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
)

const systemPrompt = `Você é um especialista em otimização de currículos com vasto conhecimento em recrutamento e seleção, ATS (Applicant Tracking Systems) e mercado de trabalho.

Sua função é reescrever currículos para maximizar a compatibilidade com a vaga desejada, respeitando o nível de senioridade e exigência da vaga.

Regras:
1. Analise o currículo original e a descrição da vaga fornecidos
2. Identifique o nível ATS da vaga (entry-level, mid-level, senior, expert)
3. Reestruture o currículo em linguagem Typst, organizando seções de forma profissional
4. Destaque palavras-chave da vaga no currículo
5. Use linguagem e profundidade condizentes com o nível ATS identificado
6. Mantenha a veracidade das informações — nunca invente experiências ou habilidades
7. Priorize realizações mensuráveis e resultados concretos
8. Otimize o formato para ser legível tanto por humanos quanto por sistemas ATS

Retorne APENAS o código Typst, sem explicações adicionais.`

type OptimizationServices struct {
	OptRepo    *repositories.OptimizationRepository
	ResumeRepo *repositories.ResumeRepository
	JobRepo    *repositories.JobRepository
	Gemini     *GeminiClient
}

func NewOptimizationServices(
	optRepo *repositories.OptimizationRepository,
	resumeRepo *repositories.ResumeRepository,
	jobRepo *repositories.JobRepository,
	gemini *GeminiClient,
) *OptimizationServices {
	return &OptimizationServices{
		OptRepo:    optRepo,
		ResumeRepo: resumeRepo,
		JobRepo:    jobRepo,
		Gemini:     gemini,
	}
}

func (s *OptimizationServices) Optimize(userID, resumeID, jobID string) (responses.OptimizeResponse, error) {
	resume, err := s.ResumeRepo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.OptimizeResponse{}, errors.New("currículo não encontrado")
		}
		return responses.OptimizeResponse{}, err
	}

	if resume.UserID.String() != userID {
		return responses.OptimizeResponse{}, errors.New("currículo não encontrado")
	}

	job, err := s.JobRepo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.OptimizeResponse{}, errors.New("vaga não encontrada")
		}
		return responses.OptimizeResponse{}, err
	}

	if job.UserID.String() != userID {
		return responses.OptimizeResponse{}, errors.New("vaga não encontrada")
	}

	userPrompt := "Currículo:\n" + resume.RawText + "\n\nDescrição da Vaga:\n" + job.RawDescription

	rawText, err := s.Gemini.SendPrompt(systemPrompt, userPrompt)
	if err != nil {
		return responses.OptimizeResponse{}, err
	}

	typstContent := extractTypstContent(rawText)

	opt := entities.ResumeOptimized{
		ID:           uuid.New(),
		ResumeID:     resume.ID,
		JobID:        job.ID,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		RawText:      rawText,
		TypstContent: typstContent,
		CreatedAt:    time.Now(),
	}

	err = s.OptRepo.Create(opt)
	if err != nil {
		return responses.OptimizeResponse{}, err
	}

	return s.toResponse(opt), nil
}

func (s *OptimizationServices) GetByResumeID(userID, resumeID string) ([]responses.OptimizeSummaryResponse, error) {
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

	opts, err := s.OptRepo.GetByResumeID(resumeID)
	if err != nil {
		return nil, err
	}

	result := make([]responses.OptimizeSummaryResponse, 0, len(opts))
	for _, opt := range opts {
		result = append(result, responses.OptimizeSummaryResponse{
			ID:        opt.ID.String(),
			ResumeID:  opt.ResumeID.String(),
			JobID:     opt.JobID.String(),
			CreatedAt: opt.CreatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}

func (s *OptimizationServices) GetByID(userID, optimizationID string) (responses.OptimizeResponse, error) {
	opt, err := s.OptRepo.GetByID(optimizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.OptimizeResponse{}, errors.New("otimização não encontrada")
		}
		return responses.OptimizeResponse{}, err
	}

	resume, err := s.ResumeRepo.GetByID(opt.ResumeID.String())
	if err != nil {
		return responses.OptimizeResponse{}, errors.New("otimização não encontrada")
	}

	if resume.UserID.String() != userID {
		return responses.OptimizeResponse{}, errors.New("otimização não encontrada")
	}

	return s.toResponse(opt), nil
}

func (s *OptimizationServices) toResponse(opt entities.ResumeOptimized) responses.OptimizeResponse {
	return responses.OptimizeResponse{
		ID:           opt.ID.String(),
		ResumeID:     opt.ResumeID.String(),
		JobID:        opt.JobID.String(),
		TypstContent: opt.TypstContent,
		CreatedAt:    opt.CreatedAt.Format(time.RFC3339),
	}
}

func extractTypstContent(raw string) string {
	raw = strings.TrimSpace(raw)

	raw = strings.TrimPrefix(raw, "```typst")
	raw = strings.TrimPrefix(raw, "```")

	raw = strings.TrimSuffix(raw, "```")

	return strings.TrimSpace(raw)
}
