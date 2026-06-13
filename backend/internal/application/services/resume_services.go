package services

import (
	"database/sql"
	"errors"
	"io"
	"time"

	"backend/internal/application/responses"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
)

type ResumeServices struct {
	Repo      *repositories.ResumeRepository
	Extractor *TextExtractor
}

func NewResumeServices(repo *repositories.ResumeRepository, extractor *TextExtractor) *ResumeServices {
	return &ResumeServices{
		Repo:      repo,
		Extractor: extractor,
	}
}

func (s *ResumeServices) Create(userID, originalName string, file io.Reader) (responses.ResumeResponse, error) {
	rawText, err := s.Extractor.ExtractText(originalName, file)
	if err != nil {
		return responses.ResumeResponse{}, err
	}

	resume := entities.Resume{
		ID:           uuid.New(),
		UserID:       uuid.MustParse(userID),
		OriginalName: originalName,
		RawText:      rawText,
		UploadedAt:   time.Now(),
	}

	err = s.Repo.Create(resume)
	if err != nil {
		return responses.ResumeResponse{}, err
	}

	return s.toResponse(resume), nil
}

func (s *ResumeServices) GetByID(userID, resumeID string) (responses.ResumeResponse, error) {
	resume, err := s.Repo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.ResumeResponse{}, errors.New("currículo não encontrado")
		}
		return responses.ResumeResponse{}, err
	}

	if resume.UserID.String() != userID {
		return responses.ResumeResponse{}, errors.New("currículo não encontrado")
	}

	return s.toResponse(resume), nil
}

func (s *ResumeServices) GetByUserID(userID string) ([]responses.ResumeSummaryResponse, error) {
	resumes, err := s.Repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]responses.ResumeSummaryResponse, 0, len(resumes))
	for _, resume := range resumes {
		result = append(result, responses.ResumeSummaryResponse{
			ID:           resume.ID,
			UserID:       resume.UserID,
			OriginalName: resume.OriginalName,
			UploadedAt:   resume.UploadedAt,
		})
	}

	return result, nil
}

func (s *ResumeServices) Update(userID, resumeID, originalName string, file io.Reader) (responses.ResumeResponse, error) {
	resume, err := s.Repo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.ResumeResponse{}, errors.New("currículo não encontrado")
		}
		return responses.ResumeResponse{}, err
	}

	if resume.UserID.String() != userID {
		return responses.ResumeResponse{}, errors.New("currículo não encontrado")
	}

	rawText, err := s.Extractor.ExtractText(originalName, file)
	if err != nil {
		return responses.ResumeResponse{}, err
	}

	resume.OriginalName = originalName
	resume.RawText = rawText
	resume.UploadedAt = time.Now()

	err = s.Repo.Update(resume)
	if err != nil {
		return responses.ResumeResponse{}, err
	}

	return s.toResponse(resume), nil
}

func (s *ResumeServices) Delete(userID, resumeID string) error {
	resume, err := s.Repo.GetByID(resumeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("currículo não encontrado")
		}
		return err
	}

	if resume.UserID.String() != userID {
		return errors.New("currículo não encontrado")
	}

	return s.Repo.Delete(resumeID)
}

func (s *ResumeServices) toResponse(resume entities.Resume) responses.ResumeResponse {
	return responses.ResumeResponse{
		ID:           resume.ID,
		UserID:       resume.UserID,
		OriginalName: resume.OriginalName,
		RawText:      resume.RawText,
		UploadedAt:   resume.UploadedAt,
	}
}
