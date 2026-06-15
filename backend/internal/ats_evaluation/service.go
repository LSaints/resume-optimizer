package atsevaluation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/job"
	"backend/internal/resume"
	"backend/pkg/ai"

	"github.com/google/uuid"
)

const atsScoringSystemPrompt = `Você é um especialista sênior em recrutamento, seleção, ATS (Applicant Tracking Systems), aquisição de talentos e análise de currículos.

Sua função é avaliar objetivamente a compatibilidade entre um currículo e uma vaga de emprego.

Você deve agir como um sistema ATS moderno utilizado por plataformas como Gupy, Greenhouse, Lever, Workday, Ashby e similares.

# Objetivo

Analise o currículo em relação à vaga e gere uma pontuação de compatibilidade ATS baseada em critérios mensuráveis.

A avaliação deve considerar exclusivamente informações presentes no currículo e na vaga.

Não invente experiências, habilidades ou requisitos.

# Metodologia de Avaliação

Calcule a nota final utilizando os pesos abaixo:

## 1. Correspondência de Palavras-chave (30%)

Avalie:

* Tecnologias
* Ferramentas
* Frameworks
* Linguagens
* Metodologias
* Certificações
* Termos técnicos

Pontuação máxima: 3.0

## 2. Compatibilidade Técnica (25%)

Avalie:

* Hard skills exigidas
* Conhecimentos técnicos desejados
* Ferramentas utilizadas
* Experiências relacionadas

Pontuação máxima: 2.5

## 3. Experiência Profissional (20%)

Avalie:

* Nível de senioridade
* Tempo de experiência compatível
* Histórico profissional
* Complexidade das responsabilidades

Pontuação máxima: 2.0

## 4. Impacto e Resultados (15%)

Avalie:

* Métricas
* Indicadores
* Resultados mensuráveis
* Conquistas profissionais

Pontuação máxima: 1.5

## 5. Estrutura e Legibilidade ATS (10%)

Avalie:

* Clareza
* Organização
* Seções bem definidas
* Facilidade de leitura por ATS

Pontuação máxima: 1.0

# Interpretação da Nota

0.0 – 2.9

Compatibilidade muito baixa.
O currículo possui pouca aderência aos requisitos da vaga.

3.0 – 4.9

Compatibilidade baixa.
Existem lacunas significativas entre o currículo e a vaga.

5.0 – 6.9

Compatibilidade moderada.
O candidato atende parte relevante dos requisitos.

7.0 – 8.4

Boa compatibilidade.
O currículo demonstra forte alinhamento com a vaga.

8.5 – 10.0

Excelente compatibilidade.
O currículo apresenta aderência muito alta aos requisitos.

# Diretrizes para Análise

Identifique:

* Principais palavras-chave da vaga
* Palavras-chave encontradas no currículo
* Competências ausentes
* Requisitos parcialmente atendidos
* Requisitos totalmente atendidos
* Possíveis melhorias para aumentar a compatibilidade ATS

Seja específico.

Evite comentários genéricos.

Sempre explique os motivos da pontuação atribuída.

# Regras

* Não invente informações.
* Não assuma conhecimentos não mencionados.
* Não penalize o candidato por requisitos classificados como desejáveis quando não forem obrigatórios.
* Considere sinônimos e tecnologias equivalentes quando apropriado.
* Considere contexto profissional ao avaliar compatibilidade.
* Utilize linguagem profissional em português brasileiro.

# Formato de Saída

Retorne APENAS um JSON válido.

Não utilize markdown.

Não utilize blocos de código.

Não adicione comentários.

O JSON deve seguir exatamente esta estrutura:

{
"score": 8.3,
"summary": "Resumo executivo da compatibilidade ATS.",
"details": "Análise detalhada dos critérios avaliados, pontos fortes, lacunas identificadas e oportunidades de melhoria.",
"breakdown": {
"keywordMatch": 2.6,
"technicalCompatibility": 2.1,
"professionalExperience": 1.8,
"impactAndResults": 1.2,
"atsReadability": 0.8
},
"matchedKeywords": [
"Palavra-chave 1",
"Palavra-chave 2"
],
"missingKeywords": [
"Palavra-chave ausente 1",
"Palavra-chave ausente 2"
],
"recommendations": [
"Recomendação 1",
"Recomendação 2"
]
}

# Validação Final

Antes de responder, confirme:

* O JSON é válido.
* O score está entre 0.0 e 10.0.
* O score possui no máximo uma casa decimal.
* Os subtotais respeitam os pesos definidos.
* Nenhuma informação foi inventada.
* Nenhum texto foi retornado fora do JSON.
`

type atsBreakdown struct {
	KeywordMatch           float64 `json:"keywordMatch"`
	TechnicalCompatibility float64 `json:"technicalCompatibility"`
	ProfessionalExperience float64 `json:"professionalExperience"`
	ImpactAndResults       float64 `json:"impactAndResults"`
	AtsReadability         float64 `json:"atsReadability"`
}

type atsScoreResponse struct {
	Score           float64      `json:"score"`
	Summary         string       `json:"summary"`
	Details         string       `json:"details"`
	Breakdown       atsBreakdown `json:"breakdown"`
	MatchedKeywords []string     `json:"matchedKeywords"`
	MissingKeywords []string     `json:"missingKeywords"`
	Recommendations []string     `json:"recommendations"`
}

func parseEvaluationResponse(raw string) (*atsScoreResponse, error) {
	raw = strings.TrimSpace(raw)

	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var resp atsScoreResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, errors.New("resposta inválida da IA")
	}

	if resp.Score < 0 || resp.Score > 10 {
		return nil, fmt.Errorf("pontuação inválida: %.1f", resp.Score)
	}

	if resp.Summary == "" || resp.Details == "" {
		return nil, errors.New("resposta inválida da IA")
	}

	if resp.Breakdown.KeywordMatch < 0 || resp.Breakdown.KeywordMatch > 3.0 {
		return nil, fmt.Errorf("breakdownKeywordMatch inválido: %.2f", resp.Breakdown.KeywordMatch)
	}
	if resp.Breakdown.TechnicalCompatibility < 0 || resp.Breakdown.TechnicalCompatibility > 2.5 {
		return nil, fmt.Errorf("breakdownTechnical inválido: %.2f", resp.Breakdown.TechnicalCompatibility)
	}
	if resp.Breakdown.ProfessionalExperience < 0 || resp.Breakdown.ProfessionalExperience > 2.0 {
		return nil, fmt.Errorf("breakdownExperience inválido: %.2f", resp.Breakdown.ProfessionalExperience)
	}
	if resp.Breakdown.ImpactAndResults < 0 || resp.Breakdown.ImpactAndResults > 1.5 {
		return nil, fmt.Errorf("breakdownImpact inválido: %.2f", resp.Breakdown.ImpactAndResults)
	}
	if resp.Breakdown.AtsReadability < 0 || resp.Breakdown.AtsReadability > 1.0 {
		return nil, fmt.Errorf("breakdownReadability inválido: %.2f", resp.Breakdown.AtsReadability)
	}

	return &resp, nil
}

type AtsScoringServices struct {
	EvalRepo   *AtsEvaluationRepository
	ResumeRepo *resume.ResumeRepository
	JobRepo    *job.JobRepository
	Gemini     *ai.GeminiClient
}

func NewAtsScoringServices(
	evalRepo *AtsEvaluationRepository,
	resumeRepo *resume.ResumeRepository,
	jobRepo *job.JobRepository,
	gemini *ai.GeminiClient,
) *AtsScoringServices {
	return &AtsScoringServices{
		EvalRepo:   evalRepo,
		ResumeRepo: resumeRepo,
		JobRepo:    jobRepo,
		Gemini:     gemini,
	}
}

func (s *AtsScoringServices) Evaluate(userID, resumeID, jobID string) (AtsEvaluationResponse, error) {
	resume, err := s.ResumeRepo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AtsEvaluationResponse{}, errors.New("currículo não encontrado")
		}
		return AtsEvaluationResponse{}, err
	}

	if resume.UserID.String() != userID {
		return AtsEvaluationResponse{}, errors.New("currículo não encontrado")
	}

	job, err := s.JobRepo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AtsEvaluationResponse{}, errors.New("vaga não encontrada")
		}
		return AtsEvaluationResponse{}, err
	}

	if job.UserID.String() != userID {
		return AtsEvaluationResponse{}, errors.New("vaga não encontrada")
	}

	userPrompt := "Currículo:\n" + resume.RawText + "\n\nDescrição da Vaga:\n" + job.RawDescription

	rawText, err := s.Gemini.SendPrompt(atsScoringSystemPrompt, userPrompt)
	if err != nil {
		return AtsEvaluationResponse{}, err
	}

	parsed, err := parseEvaluationResponse(rawText)
	if err != nil {
		return AtsEvaluationResponse{}, err
	}

	matchedKeywordsJSON, _ := json.Marshal(parsed.MatchedKeywords)
	missingKeywordsJSON, _ := json.Marshal(parsed.MissingKeywords)
	recommendationsJSON, _ := json.Marshal(parsed.Recommendations)

	eval := AtsEvaluation{
		ID:                    uuid.New(),
		ResumeID:              resume.ID,
		JobID:                 job.ID,
		Score:                 parsed.Score,
		Summary:               parsed.Summary,
		Details:               parsed.Details,
		RawResponse:           rawText,
		BreakdownKeywordMatch: parsed.Breakdown.KeywordMatch,
		BreakdownTechnical:    parsed.Breakdown.TechnicalCompatibility,
		BreakdownExperience:   parsed.Breakdown.ProfessionalExperience,
		BreakdownImpact:       parsed.Breakdown.ImpactAndResults,
		BreakdownReadability:  parsed.Breakdown.AtsReadability,
		MatchedKeywords:       string(matchedKeywordsJSON),
		MissingKeywords:       string(missingKeywordsJSON),
		Recommendations:       string(recommendationsJSON),
		CreatedAt:             time.Now(),
	}

	err = s.EvalRepo.Create(eval)
	if err != nil {
		return AtsEvaluationResponse{}, err
	}

	return s.toResponse(eval), nil
}

func (s *AtsScoringServices) ListByResume(userID, resumeID string) ([]AtsEvaluationSummaryResponse, error) {
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

	result := make([]AtsEvaluationSummaryResponse, 0, len(evals))
	for _, eval := range evals {
		result = append(result, AtsEvaluationSummaryResponse{
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

func (s *AtsScoringServices) GetByID(userID, evaluationID string) (AtsEvaluationResponse, error) {
	eval, err := s.EvalRepo.GetByID(evaluationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
		}
		return AtsEvaluationResponse{}, err
	}

	resume, err := s.ResumeRepo.GetByID(eval.ResumeID.String())
	if err != nil {
		return AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
	}

	if resume.UserID.String() != userID {
		return AtsEvaluationResponse{}, errors.New("avaliação não encontrada")
	}

	return s.toResponse(eval), nil
}

func (s *AtsScoringServices) toResponse(eval AtsEvaluation) AtsEvaluationResponse {
	matchedKeywords := deserializeStringSlice(eval.MatchedKeywords)
	missingKeywords := deserializeStringSlice(eval.MissingKeywords)
	recommendations := deserializeStringSlice(eval.Recommendations)

	return AtsEvaluationResponse{
		ID:                    eval.ID.String(),
		ResumeID:              eval.ResumeID.String(),
		JobID:                 eval.JobID.String(),
		Score:                 eval.Score,
		Summary:               eval.Summary,
		Details:               eval.Details,
		BreakdownKeywordMatch: eval.BreakdownKeywordMatch,
		BreakdownTechnical:    eval.BreakdownTechnical,
		BreakdownExperience:   eval.BreakdownExperience,
		BreakdownImpact:       eval.BreakdownImpact,
		BreakdownReadability:  eval.BreakdownReadability,
		MatchedKeywords:       matchedKeywords,
		MissingKeywords:       missingKeywords,
		Recommendations:       recommendations,
		CreatedAt:             eval.CreatedAt.Format(time.RFC3339),
	}
}

func deserializeStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return []string{}
	}
	if result == nil {
		return []string{}
	}
	return result
}
