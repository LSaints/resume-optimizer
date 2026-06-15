package render

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

var (
	reBullet     = regexp.MustCompile(`•`)
	reStarDot    = regexp.MustCompile(`(^|[^\\])\*\.`)
	reStarBullet = regexp.MustCompile(`(?m)^\*\s+`)
	reAtSign     = regexp.MustCompile(`(^|[^\\])@`)
)

func sanitizeTypst(input string) string {
	input = reBullet.ReplaceAllString(input, "-")
	input = reStarBullet.ReplaceAllString(input, "- ")
	input = reStarDot.ReplaceAllString(input, "$1\\*.")
	input = reAtSign.ReplaceAllString(input, "$1\\@")
	return input
}

type TypstRenderService struct{}

func NewTypstRenderService() *TypstRenderService {
	return &TypstRenderService{}
}

func (s *TypstRenderService) RenderToSVG(typstContent string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return s.render(ctx, typstContent, "svg")
}

func (s *TypstRenderService) RenderToPDF(typstContent string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := s.render(ctx, typstContent, "pdf")
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

func (s *TypstRenderService) render(ctx context.Context, typstContent, format string) (string, error) {
	if _, err := exec.LookPath("typst"); err != nil {
		return "", fmt.Errorf("renderizador typst nao disponivel")
	}

	dir, err := os.MkdirTemp("", "typst-render-*")
	if err != nil {
		return "", fmt.Errorf("erro ao criar diretorio temporario: %w", err)
	}
	defer os.RemoveAll(dir)

	sanitized := sanitizeTypst(typstContent)

	inputPath := filepath.Join(dir, "input.typ")
	if err := os.WriteFile(inputPath, []byte(sanitized), 0644); err != nil {
		return "", fmt.Errorf("erro ao criar arquivo temporario: %w", err)
	}

	outputName := "output." + format
	outputPath := filepath.Join(dir, outputName)

	args := []string{"compile", inputPath, outputPath}
	if format == "svg" {
		args = append(args, "--format", "svg")
	}

	cmd := exec.CommandContext(ctx, "typst", args...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao renderizar documento: %s", string(out))
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo renderizado: %w", err)
	}

	return string(data), nil
}
