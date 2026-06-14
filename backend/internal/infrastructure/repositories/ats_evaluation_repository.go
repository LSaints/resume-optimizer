package repositories

import (
	"backend/internal/domain/entities"
	"database/sql"
)

type AtsEvaluationRepository struct {
	db *sql.DB
}

func NewAtsEvaluationRepository(db *sql.DB) *AtsEvaluationRepository {
	return &AtsEvaluationRepository{
		db: db,
	}
}

func (r *AtsEvaluationRepository) Create(eval entities.AtsEvaluation) error {
	query := `
		INSERT INTO ats_evaluations (id, resume_id, job_id, score, summary, details, raw_response, breakdown_keyword_match, breakdown_technical, breakdown_experience, breakdown_impact, breakdown_readability, matched_keywords, missing_keywords, recommendations, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		eval.ID,
		eval.ResumeID,
		eval.JobID,
		eval.Score,
		eval.Summary,
		eval.Details,
		eval.RawResponse,
		eval.BreakdownKeywordMatch,
		eval.BreakdownTechnical,
		eval.BreakdownExperience,
		eval.BreakdownImpact,
		eval.BreakdownReadability,
		eval.MatchedKeywords,
		eval.MissingKeywords,
		eval.Recommendations,
		eval.CreatedAt,
	)

	return err
}

func (r *AtsEvaluationRepository) GetByID(id string) (entities.AtsEvaluation, error) {
	query := `
		SELECT id, resume_id, job_id, score, summary, details, raw_response, breakdown_keyword_match, breakdown_technical, breakdown_experience, breakdown_impact, breakdown_readability, matched_keywords, missing_keywords, recommendations, created_at
		FROM ats_evaluations
		WHERE id = ?
	`

	var eval entities.AtsEvaluation

	err := r.db.QueryRow(query, id).Scan(
		&eval.ID,
		&eval.ResumeID,
		&eval.JobID,
		&eval.Score,
		&eval.Summary,
		&eval.Details,
		&eval.RawResponse,
		&eval.BreakdownKeywordMatch,
		&eval.BreakdownTechnical,
		&eval.BreakdownExperience,
		&eval.BreakdownImpact,
		&eval.BreakdownReadability,
		&eval.MatchedKeywords,
		&eval.MissingKeywords,
		&eval.Recommendations,
		&eval.CreatedAt,
	)

	if err != nil {
		return entities.AtsEvaluation{}, err
	}

	return eval, nil
}

func (r *AtsEvaluationRepository) GetByResumeID(resumeID string) ([]entities.AtsEvaluation, error) {
	query := `
		SELECT id, resume_id, job_id, score, summary, details, raw_response, breakdown_keyword_match, breakdown_technical, breakdown_experience, breakdown_impact, breakdown_readability, matched_keywords, missing_keywords, recommendations, created_at
		FROM ats_evaluations
		WHERE resume_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, resumeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evals []entities.AtsEvaluation

	for rows.Next() {
		var eval entities.AtsEvaluation

		err := rows.Scan(
			&eval.ID,
			&eval.ResumeID,
			&eval.JobID,
			&eval.Score,
			&eval.Summary,
			&eval.Details,
			&eval.RawResponse,
			&eval.BreakdownKeywordMatch,
			&eval.BreakdownTechnical,
			&eval.BreakdownExperience,
			&eval.BreakdownImpact,
			&eval.BreakdownReadability,
			&eval.MatchedKeywords,
			&eval.MissingKeywords,
			&eval.Recommendations,
			&eval.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		evals = append(evals, eval)
	}

	if evals == nil {
		return []entities.AtsEvaluation{}, nil
	}

	return evals, nil
}

func (r *AtsEvaluationRepository) Delete(id string) error {
	query := `
		DELETE FROM ats_evaluations
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
