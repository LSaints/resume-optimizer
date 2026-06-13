package services

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/application/requests"
	"backend/internal/application/responses"
	"backend/internal/domain/entities"
	"backend/internal/infrastructure/repositories"

	"github.com/google/uuid"
)

type JobServices struct {
	Repo *repositories.JobRepository
}

func NewJobServices(repo *repositories.JobRepository) *JobServices {
	return &JobServices{Repo: repo}
}

func (s *JobServices) Create(userID string, request requests.CreateJobRequest) (responses.JobResponse, error) {
	if request.Title == "" {
		return responses.JobResponse{}, errors.New("título é obrigatório")
	}

	if request.RawDescription == "" {
		return responses.JobResponse{}, errors.New("descrição é obrigatória")
	}

	now := time.Now()

	job := entities.Job{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		Title:          request.Title,
		RawDescription: request.RawDescription,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.Repo.Create(job)
	if err != nil {
		return responses.JobResponse{}, err
	}

	return s.toResponse(job), nil
}

func (s *JobServices) GetByID(userID, jobID string) (responses.JobResponse, error) {
	job, err := s.Repo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.JobResponse{}, errors.New("vaga não encontrada")
		}
		return responses.JobResponse{}, err
	}

	if job.UserID.String() != userID {
		return responses.JobResponse{}, errors.New("vaga não encontrada")
	}

	return s.toResponse(job), nil
}

func (s *JobServices) GetByUserID(userID string) ([]responses.JobResponse, error) {
	jobs, err := s.Repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]responses.JobResponse, 0, len(jobs))
	for _, job := range jobs {
		result = append(result, s.toResponse(job))
	}

	return result, nil
}

func (s *JobServices) Update(userID, jobID string, request requests.UpdateJobRequest) (responses.JobResponse, error) {
	if request.Title == "" {
		return responses.JobResponse{}, errors.New("título é obrigatório")
	}

	if request.RawDescription == "" {
		return responses.JobResponse{}, errors.New("descrição é obrigatória")
	}

	job, err := s.Repo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.JobResponse{}, errors.New("vaga não encontrada")
		}
		return responses.JobResponse{}, err
	}

	if job.UserID.String() != userID {
		return responses.JobResponse{}, errors.New("vaga não encontrada")
	}

	job.Title = request.Title
	job.RawDescription = request.RawDescription
	job.UpdatedAt = time.Now()

	err = s.Repo.Update(job)
	if err != nil {
		return responses.JobResponse{}, err
	}

	return s.toResponse(job), nil
}

func (s *JobServices) Delete(userID, jobID string) error {
	job, err := s.Repo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("vaga não encontrada")
		}
		return err
	}

	if job.UserID.String() != userID {
		return errors.New("vaga não encontrada")
	}

	return s.Repo.Delete(jobID)
}

func (s *JobServices) toResponse(job entities.Job) responses.JobResponse {
	return responses.JobResponse{
		ID:             job.ID,
		Title:          job.Title,
		RawDescription: job.RawDescription,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}
}
