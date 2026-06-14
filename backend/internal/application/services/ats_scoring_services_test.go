package services

import (
	"math"
	"testing"
)

func TestParseEvaluationResponse_Completo(t *testing.T) {
	raw := `{
		"score": 8.3,
		"summary": "Resumo executivo",
		"details": "Análise detalhada",
		"breakdown": {
			"keywordMatch": 2.6,
			"technicalCompatibility": 2.1,
			"professionalExperience": 1.6,
			"impactAndResults": 1.2,
			"atsReadability": 0.8
		},
		"matchedKeywords": ["Go", "SQL"],
		"missingKeywords": ["Python"],
		"recommendations": ["Adicione Python"]
	}`

	resp, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Score != 8.3 {
		t.Fatalf("score: esperado 8.3, got %.1f", resp.Score)
	}
	if resp.Summary != "Resumo executivo" {
		t.Fatalf("summary incorreto")
	}
	if resp.Details != "Análise detalhada" {
		t.Fatalf("details incorreto")
	}

	b := resp.Breakdown
	if b.KeywordMatch != 2.6 {
		t.Fatalf("keywordMatch: esperado 2.6, got %.1f", b.KeywordMatch)
	}
	if b.TechnicalCompatibility != 2.1 {
		t.Fatalf("technicalCompatibility: esperado 2.1, got %.1f", b.TechnicalCompatibility)
	}
	if b.ProfessionalExperience != 1.6 {
		t.Fatalf("professionalExperience: esperado 1.6, got %.1f", b.ProfessionalExperience)
	}
	if b.ImpactAndResults != 1.2 {
		t.Fatalf("impactAndResults: esperado 1.2, got %.1f", b.ImpactAndResults)
	}
	if b.AtsReadability != 0.8 {
		t.Fatalf("atsReadability: esperado 0.8, got %.1f", b.AtsReadability)
	}

	if len(resp.MatchedKeywords) != 2 || resp.MatchedKeywords[0] != "Go" {
		t.Fatalf("matchedKeywords incorreto: %v", resp.MatchedKeywords)
	}
	if len(resp.MissingKeywords) != 1 || resp.MissingKeywords[0] != "Python" {
		t.Fatalf("missingKeywords incorreto: %v", resp.MissingKeywords)
	}
	if len(resp.Recommendations) != 1 || resp.Recommendations[0] != "Adicione Python" {
		t.Fatalf("recommendations incorreto: %v", resp.Recommendations)
	}

	soma := b.KeywordMatch + b.TechnicalCompatibility + b.ProfessionalExperience + b.ImpactAndResults + b.AtsReadability
	if math.Abs(soma-resp.Score) > 0.1 {
		t.Fatalf("soma dos breakdowns (%.1f) difere do score (%.1f)", soma, resp.Score)
	}
}

func TestParseEvaluationResponse_SemBreakdownEListas(t *testing.T) {
	raw := `{
		"score": 7.0,
		"summary": "Resumo",
		"details": "Detalhes"
	}`

	resp, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Score != 7.0 {
		t.Fatalf("score incorreto")
	}

	b := resp.Breakdown
	if b.KeywordMatch != 0 || b.TechnicalCompatibility != 0 || b.ProfessionalExperience != 0 || b.ImpactAndResults != 0 || b.AtsReadability != 0 {
		t.Fatal("breakdown deveria ser zero")
	}

	if resp.MatchedKeywords != nil {
		t.Fatal("matchedKeywords deveria ser nil")
	}
	if resp.MissingKeywords != nil {
		t.Fatal("missingKeywords deveria ser nil")
	}
	if resp.Recommendations != nil {
		t.Fatal("recommendations deveria ser nil")
	}
}

func TestParseEvaluationResponse_JSONParcial(t *testing.T) {
	raw := `{
		"score": 6.5,
		"summary": "Resumo",
		"details": "Detalhes",
		"breakdown": {
			"keywordMatch": 2.0,
			"technicalCompatibility": 1.5,
			"professionalExperience": 1.0,
			"impactAndResults": 0.5,
			"atsReadability": 0.5
		}
	}`

	resp, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Breakdown.KeywordMatch != 2.0 {
		t.Fatalf("keywordMatch incorreto")
	}
	if resp.MatchedKeywords != nil {
		t.Fatal("matchedKeywords deveria ser nil quando ausente")
	}
}

func TestParseEvaluationResponse_ScoreInvalido(t *testing.T) {
	raw := `{"score": 15, "summary": "x", "details": "y"}`

	_, err := parseEvaluationResponse(raw)
	if err == nil {
		t.Fatal("esperava erro para score > 10")
	}
}

func TestParseEvaluationResponse_SummaryVazio(t *testing.T) {
	raw := `{"score": 5, "summary": "", "details": "y"}`

	_, err := parseEvaluationResponse(raw)
	if err == nil {
		t.Fatal("esperava erro para summary vazio")
	}
}

func TestParseEvaluationResponse_JSONInvalido(t *testing.T) {
	raw := `{invalid}`

	_, err := parseEvaluationResponse(raw)
	if err == nil {
		t.Fatal("esperava erro para JSON inválido")
	}
}

func TestParseEvaluationResponse_ComCodeBlock(t *testing.T) {
	raw := "```json\n{\"score\": 7.5, \"summary\": \"Resumo\", \"details\": \"Detalhes\"}\n```"

	resp, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Score != 7.5 {
		t.Fatalf("score incorreto")
	}
}

func TestParseEvaluationResponse_QuebraDeLinha(t *testing.T) {
	raw := "{\n\"score\": 5.0,\n\"summary\": \"Resumo\",\n\"details\": \"Detalhes\"\n}"

	resp, err := parseEvaluationResponse(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Score != 5.0 {
		t.Fatalf("score incorreto")
	}
}

func TestParseEvaluationResponse_BreakdownExcedeMaximo(t *testing.T) {
	raw := `{
		"score": 9.0,
		"summary": "Resumo",
		"details": "Detalhes",
		"breakdown": {
			"keywordMatch": 3.5,
			"technicalCompatibility": 2.0,
			"professionalExperience": 1.0,
			"impactAndResults": 0.5,
			"atsReadability": 0.5
		}
	}`

	_, err := parseEvaluationResponse(raw)
	if err == nil {
		t.Fatal("esperava erro para breakdownKeywordMatch > 3.0")
	}
}

func TestParseEvaluationResponse_BreakdownNegativo(t *testing.T) {
	raw := `{
		"score": 5.0,
		"summary": "Resumo",
		"details": "Detalhes",
		"breakdown": {
			"keywordMatch": -1.0,
			"technicalCompatibility": 2.0,
			"professionalExperience": 1.0,
			"impactAndResults": 0.5,
			"atsReadability": 0.5
		}
	}`

	_, err := parseEvaluationResponse(raw)
	if err == nil {
		t.Fatal("esperava erro para breakdownKeywordMatch < 0")
	}
}
