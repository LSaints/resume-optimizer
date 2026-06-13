package repositories

import (
	"backend/internal/domain/entities"
	"database/sql"
)

type JobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) Create(job entities.Job) error {
	query := `
		INSERT INTO jobs (id, user_id, title, raw_description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		job.ID,
		job.UserID,
		job.Title,
		job.RawDescription,
		job.CreatedAt,
		job.UpdatedAt,
	)

	return err
}

func (r *JobRepository) GetByID(id string) (entities.Job, error) {
	query := `
		SELECT id, user_id, title, raw_description, created_at, updated_at
		FROM jobs
		WHERE id = ?
	`

	var job entities.Job

	err := r.db.QueryRow(query, id).Scan(
		&job.ID,
		&job.UserID,
		&job.Title,
		&job.RawDescription,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		return entities.Job{}, err
	}

	return job, nil
}

func (r *JobRepository) GetByUserID(userID string) ([]entities.Job, error) {
	query := `
		SELECT id, user_id, title, raw_description, created_at, updated_at
		FROM jobs
		WHERE user_id = ?
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []entities.Job

	for rows.Next() {
		var job entities.Job

		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.Title,
			&job.RawDescription,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	if jobs == nil {
		return []entities.Job{}, nil
	}

	return jobs, nil
}

func (r *JobRepository) Update(job entities.Job) error {
	query := `
		UPDATE jobs
		SET title = ?, raw_description = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		job.Title,
		job.RawDescription,
		job.UpdatedAt,
		job.ID,
	)

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

func (r *JobRepository) Delete(id string) error {
	query := `
		DELETE FROM jobs
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
