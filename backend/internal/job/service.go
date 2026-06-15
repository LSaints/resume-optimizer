package job

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type JobServices struct {
	Repo *JobRepository
}

func NewJobServices(repo *JobRepository) *JobServices {
	return &JobServices{Repo: repo}
}

func (s *JobServices) Create(userID string, request CreateJobRequest) (JobResponse, error) {
	if request.Title == "" {
		return JobResponse{}, errors.New("título é obrigatório")
	}

	if request.RawDescription == "" {
		return JobResponse{}, errors.New("descrição é obrigatória")
	}

	now := time.Now()

	job := Job{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		Title:          request.Title,
		RawDescription: request.RawDescription,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.Repo.Create(job)
	if err != nil {
		return JobResponse{}, err
	}

	return s.toResponse(job), nil
}

func (s *JobServices) GetByID(userID, jobID string) (JobResponse, error) {
	job, err := s.Repo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobResponse{}, errors.New("vaga não encontrada")
		}
		return JobResponse{}, err
	}

	if job.UserID.String() != userID {
		return JobResponse{}, errors.New("vaga não encontrada")
	}

	return s.toResponse(job), nil
}

func (s *JobServices) GetByUserID(userID string) ([]JobResponse, error) {
	jobs, err := s.Repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]JobResponse, 0, len(jobs))
	for _, job := range jobs {
		result = append(result, s.toResponse(job))
	}

	return result, nil
}

func (s *JobServices) Update(userID, jobID string, request UpdateJobRequest) (JobResponse, error) {
	if request.Title == "" {
		return JobResponse{}, errors.New("título é obrigatório")
	}

	if request.RawDescription == "" {
		return JobResponse{}, errors.New("descrição é obrigatória")
	}

	job, err := s.Repo.GetByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return JobResponse{}, errors.New("vaga não encontrada")
		}
		return JobResponse{}, err
	}

	if job.UserID.String() != userID {
		return JobResponse{}, errors.New("vaga não encontrada")
	}

	job.Title = request.Title
	job.RawDescription = request.RawDescription
	job.UpdatedAt = time.Now()

	err = s.Repo.Update(job)
	if err != nil {
		return JobResponse{}, err
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

func (s *JobServices) toResponse(job Job) JobResponse {
	return JobResponse{
		ID:             job.ID,
		Title:          job.Title,
		RawDescription: job.RawDescription,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}
}
