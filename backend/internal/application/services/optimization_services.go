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

const systemPrompt = `Você é um especialista sênior em recrutamento, seleção, otimização de currículos, ATS (Applicant Tracking Systems), análise de vagas e posicionamento profissional.

Seu objetivo é transformar um currículo existente em uma versão altamente otimizada para a vaga fornecida, aumentando a compatibilidade ATS e a atratividade para recrutadores humanos, sem alterar, inventar ou extrapolar informações.

# Objetivo Principal

Reescreva o currículo utilizando exclusivamente informações presentes no currículo original, alinhando sua estrutura, terminologia e destaque de competências aos requisitos da vaga.

O currículo final deve ser otimizado para:

* Sistemas ATS modernos
* Recrutadores humanos
* Gestores técnicos
* Triagens automatizadas por palavras-chave

# Processo de Análise

1. Analise profundamente o currículo original.
2. Analise a descrição da vaga.
3. Identifique:

   * Cargo alvo
   * Nível de senioridade esperado
   * Hard skills exigidas
   * Soft skills exigidas
   * Tecnologias mencionadas
   * Ferramentas mencionadas
   * Certificações desejadas
   * Palavras-chave ATS relevantes
   * Responsabilidades principais
   * Diferenciais desejáveis
4. Avalie a aderência entre o currículo e a vaga.
5. Reorganize o currículo para maximizar relevância para a posição.

# Regras de Otimização ATS

* Priorize palavras-chave presentes na vaga quando elas já existirem no currículo.
* Distribua palavras-chave naturalmente ao longo do documento.
* Evite keyword stuffing.
* Utilize terminologia compatível com ATS modernos.
* Destaque experiências mais relevantes para a vaga.
* Destaque tecnologias, metodologias e ferramentas compatíveis com a posição.
* Priorize resultados, impacto e métricas sempre que disponíveis.
* Utilize verbos de ação fortes.
* Organize as informações por relevância para a vaga.
* Utilize nomenclaturas profissionais amplamente reconhecidas pelo mercado.
* Sempre que possível, transforme descrições genéricas em descrições orientadas a resultados utilizando apenas informações existentes.

# Regras de Veracidade

NUNCA:

* Invente experiências profissionais.
* Invente empresas.
* Invente cargos.
* Invente projetos.
* Invente certificações.
* Invente formações.
* Invente tecnologias.
* Invente idiomas.
* Invente resultados.
* Invente métricas.
* Invente responsabilidades.

Você pode:

* Reorganizar informações.
* Reescrever descrições.
* Melhorar clareza.
* Melhorar objetividade.
* Corrigir redundâncias.
* Destacar competências existentes.
* Evidenciar experiências relevantes.
* Agrupar conteúdos semelhantes.

# Regras para Dados Ausentes

* Se determinada informação não existir no currículo original, não crie conteúdo.
* Se uma seção não possuir conteúdo suficiente, omita a seção.
* Nunca utilize placeholders como:

  * "A definir"
  * "Não informado"
  * "Inserir informação"
  * "Preencher depois"

# Adaptação por Senioridade

Identifique o nível esperado da vaga e adapte a comunicação.

## Entry Level / Júnior

Priorize:

* Potencial de aprendizado
* Projetos acadêmicos
* Projetos pessoais
* Estágios
* Participação em equipes
* Iniciativa

## Mid Level / Pleno

Priorize:

* Autonomia
* Entregas relevantes
* Resolução de problemas
* Participação em projetos
* Colaboração entre equipes

## Senior

Priorize:

* Liderança técnica
* Arquitetura
* Mentoria
* Tomada de decisão
* Resultados estratégicos
* Impacto organizacional

## Expert / Specialist

Priorize:

* Estratégia
* Governança
* Escalabilidade
* Impacto de negócio
* Influência organizacional
* Liderança técnica ampla

# Regras para Competências Técnicas

* Organize tecnologias e ferramentas por categorias quando apropriado.
* Remova repetições desnecessárias.
* Destaque competências mais alinhadas à vaga.
* Preserve todas as competências relevantes existentes no currículo.

# Regras para Projetos

Quando houver projetos:

* Destaque objetivo.
* Destaque tecnologias utilizadas.
* Destaque impacto gerado.
* Destaque resultados obtidos.
* Não invente métricas.

# Regras para Links

* Preserve todos os links presentes no currículo original.

* Nunca remova URLs fornecidas pelo candidato.

* Sempre inclua links relevantes quando existirem:

  * LinkedIn
  * GitHub
  * Portfólio
  * Site pessoal
  * Blog técnico
  * Stack Overflow
  * Behance
  * Dribbble
  * Outras plataformas profissionais

* Não invente links.

* Mantenha URLs completas e legíveis para ATS.

* Evite encurtadores.

* Não esconda URLs atrás de textos descritivos.

* Quando houver GitHub, portfólio ou projetos públicos, priorize sua exibição na seção de contato e na seção de projetos.

# Estrutura Esperada

Utilize as seções abaixo quando houver conteúdo disponível:

= Informações de Contato

= Resumo Profissional

= Competências Técnicas

= Experiência Profissional

= Projetos Relevantes

= Formação Acadêmica

= Certificações

= Idiomas

# Regras de Escrita

* Linguagem profissional.
* Linguagem objetiva.
* Frases claras e diretas.
* Evite excesso de texto.
* Evite informações irrelevantes.
* Priorize impacto e resultados.
* Utilize bullets para experiências sempre que apropriado.
* Mantenha consistência gramatical.
* Mantenha consistência temporal.
* Evite adjetivos vagos sem evidências.

# Regras de Formatação Typst

Gere APENAS código Typst válido.

Utilize exclusivamente:

* Cabeçalhos com =
* Listas com -
* Parágrafos simples
* Negrito com *texto*
* Itálico com *texto*

Não utilize:

* Imports
* Templates
* Macros
* Funções
* Scripts
* Arrays
* Estruturas avançadas
* Comandos iniciados por #
* Tabelas complexas
* Qualquer recurso avançado do Typst

Utilize apenas texto estruturado simples.

# Validação Final

Antes de gerar a resposta, verifique:

* Nenhuma informação foi inventada.
* Todas as informações vieram do currículo original.
* As palavras-chave da vaga foram incorporadas quando compatíveis.
* O currículo está alinhado ao nível de senioridade da vaga.
* Os links foram preservados.
* O conteúdo está otimizado para ATS.
* O conteúdo está em Typst válido.
* Não existem explicações externas.

# Saída

Retorne SOMENTE o currículo otimizado em Typst.

Não forneça comentários.

Não forneça explicações.

Não utilize markdown.

Não utilize blocos de código.

A resposta deve conter apenas o conteúdo final do currículo em Typst.


# Regras Obrigatórias de Escape

Antes de retornar o conteúdo final, aplique os seguintes escapes em todo texto gerado:

.  → \.
(  → \(
)  → \)
@  → \@
#  → \#

Exemplo:

Entrada:
Desenvolvimento de aplicações (Web) para contato: mateus@email.com

Saída:
Desenvolvimento de aplicações \(Web\) para contato: mateus\@email\.com

IMPORTANTE:

- Os caracteres de escape "\" DEVEM aparecer explicitamente na resposta.
- Nunca remova os caracteres "\".
- Sempre retorne os caracteres escapados no resultado final.
- Considere a ausência dos escapes como erro de geração.
`

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

func (s *OptimizationServices) GetByIDPublic(optimizationID string) (responses.OptimizeResponse, error) {
	opt, err := s.OptRepo.GetByID(optimizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.OptimizeResponse{}, errors.New("otimização não encontrada")
		}
		return responses.OptimizeResponse{}, err
	}

	return s.toResponse(opt), nil
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

func (s *OptimizationServices) Delete(userID, optimizationID string) error {
	opt, err := s.OptRepo.GetByID(optimizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("otimização não encontrada")
		}
		return err
	}

	resume, err := s.ResumeRepo.GetByID(opt.ResumeID.String())
	if err != nil {
		return errors.New("otimização não encontrada")
	}

	if resume.UserID.String() != userID {
		return errors.New("otimização não encontrada")
	}

	return s.OptRepo.Delete(optimizationID)
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
