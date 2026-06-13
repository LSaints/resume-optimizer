package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/nguyenthenguyen/docx"
)

type TextExtractor struct{}

func NewTextExtractor() *TextExtractor {
	return &TextExtractor{}
}

func (e *TextExtractor) ExtractText(filename string, file io.Reader) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	dotIdx := strings.LastIndex(filename, ".")
	if dotIdx == -1 {
		return "", fmt.Errorf("formato de arquivo não suportado")
	}
	ext := strings.ToLower(filename[dotIdx:])

	switch ext {
	case ".pdf":
		return e.extractPDF(data)
	case ".docx":
		return e.extractDOCX(data)
	default:
		return "", fmt.Errorf("formato de arquivo não suportado: %s", ext)
	}
}

func (e *TextExtractor) extractPDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("erro ao ler PDF: %w", err)
	}

	txtReader, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto do PDF: %w", err)
	}

	text, err := io.ReadAll(txtReader)
	if err != nil {
		return "", fmt.Errorf("erro ao ler texto extraído do PDF: %w", err)
	}

	return string(text), nil
}

func (e *TextExtractor) extractDOCX(data []byte) (string, error) {
	r, err := docx.ReadDocxFromMemory(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("erro ao ler DOCX: %w", err)
	}
	defer r.Close()

	doc := r.Editable()
	xmlContent := doc.GetContent()

	text, err := extractTextFromDocxXML(xmlContent)
	if err != nil {
		return "", fmt.Errorf("erro ao extrair texto do DOCX: %w", err)
	}

	return text, nil
}

func extractTextFromDocxXML(xmlContent string) (string, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlContent))
	var text strings.Builder
	inT := false

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inT = true
			}
		case xml.CharData:
			if inT {
				text.Write(t)
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inT = false
			}
		}
	}

	return text.String(), nil
}
