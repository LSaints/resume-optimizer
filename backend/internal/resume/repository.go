package resume

import (
	"database/sql"
)

type ResumeRepository struct {
	db *sql.DB
}

func NewResumeRepository(db *sql.DB) *ResumeRepository {
	return &ResumeRepository{
		db: db,
	}
}

func (r *ResumeRepository) Create(resume Resume) error {
	query := `
		INSERT INTO resumes (id, user_id, original_name, raw_text, uploaded_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		resume.ID,
		resume.UserID,
		resume.OriginalName,
		resume.RawText,
		resume.UploadedAt,
	)

	return err
}

func (r *ResumeRepository) GetByID(id string) (Resume, error) {
	query := `
		SELECT id, user_id, original_name, raw_text, uploaded_at
		FROM resumes
		WHERE id = ?
	`

	var resume Resume

	err := r.db.QueryRow(query, id).Scan(
		&resume.ID,
		&resume.UserID,
		&resume.OriginalName,
		&resume.RawText,
		&resume.UploadedAt,
	)

	if err != nil {
		return Resume{}, err
	}

	return resume, nil
}

func (r *ResumeRepository) GetByUserID(userID string) ([]Resume, error) {
	query := `
		SELECT id, user_id, original_name, raw_text, uploaded_at
		FROM resumes
		WHERE user_id = ?
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumes []Resume

	for rows.Next() {
		var resume Resume

		err := rows.Scan(
			&resume.ID,
			&resume.UserID,
			&resume.OriginalName,
			&resume.RawText,
			&resume.UploadedAt,
		)

		if err != nil {
			return nil, err
		}

		resumes = append(resumes, resume)
	}

	if resumes == nil {
		return []Resume{}, nil
	}

	return resumes, nil
}

func (r *ResumeRepository) Update(resume Resume) error {
	query := `
		UPDATE resumes
		SET original_name = ?, raw_text = ?, uploaded_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		resume.OriginalName,
		resume.RawText,
		resume.UploadedAt,
		resume.ID,
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

func (r *ResumeRepository) Delete(id string) error {
	query := `
		DELETE FROM resumes
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
