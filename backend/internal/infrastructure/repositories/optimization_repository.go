package repositories

import (
	"backend/internal/domain/entities"
	"database/sql"
)

type OptimizationRepository struct {
	db *sql.DB
}

func NewOptimizationRepository(db *sql.DB) *OptimizationRepository {
	return &OptimizationRepository{
		db: db,
	}
}

func (r *OptimizationRepository) Create(opt entities.ResumeOptimized) error {
	query := `
		INSERT INTO resumes_optimized (id, resume_id, job_id, system_prompt, user_prompt, raw_text, typst_content, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		opt.ID,
		opt.ResumeID,
		opt.JobID,
		opt.SystemPrompt,
		opt.UserPrompt,
		opt.RawText,
		opt.TypstContent,
		opt.CreatedAt,
	)

	return err
}

func (r *OptimizationRepository) GetByID(id string) (entities.ResumeOptimized, error) {
	query := `
		SELECT id, resume_id, job_id, system_prompt, user_prompt, raw_text, typst_content, created_at
		FROM resumes_optimized
		WHERE id = ?
	`

	var opt entities.ResumeOptimized

	err := r.db.QueryRow(query, id).Scan(
		&opt.ID,
		&opt.ResumeID,
		&opt.JobID,
		&opt.SystemPrompt,
		&opt.UserPrompt,
		&opt.RawText,
		&opt.TypstContent,
		&opt.CreatedAt,
	)

	if err != nil {
		return entities.ResumeOptimized{}, err
	}

	return opt, nil
}

func (r *OptimizationRepository) GetByResumeID(resumeID string) ([]entities.ResumeOptimized, error) {
	query := `
		SELECT id, resume_id, job_id, system_prompt, user_prompt, raw_text, typst_content, created_at
		FROM resumes_optimized
		WHERE resume_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, resumeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var opts []entities.ResumeOptimized

	for rows.Next() {
		var opt entities.ResumeOptimized

		err := rows.Scan(
			&opt.ID,
			&opt.ResumeID,
			&opt.JobID,
			&opt.SystemPrompt,
			&opt.UserPrompt,
			&opt.RawText,
			&opt.TypstContent,
			&opt.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		opts = append(opts, opt)
	}

	if opts == nil {
		return []entities.ResumeOptimized{}, nil
	}

	return opts, nil
}

func (r *OptimizationRepository) Delete(id string) error {
	query := `
		DELETE FROM resumes_optimized
		WHERE id = ?
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
