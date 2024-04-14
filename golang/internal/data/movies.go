package data

import (
	"database/sql" // New import
	"errors"
	"time"

	"github.com/lib/pq"                   // New import
	"golang.agzam.net/internal/validator" // New import
)

type ModelInfo struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ModuleName     string    `json:"module_name"`
	ModuleDuration Runtime   `json:"runtime,omitempty"`
	ExamType       []string  `json:"types,omitempty"`
	Version        int32     `json:"version"`
}

func ValidateModel(v *validator.Validator, model *ModelInfo) {
	v.Check(model.ModuleName != "", "modul_name", "must be provided")
	v.Check(len(model.ModuleName) <= 500, "modul_name", "must not be more than 500 bytes long")
	v.Check(model.ModuleDuration != 0, "runtime", "must be provided")
	v.Check(model.ModuleDuration > 0, "runtime", "must be a positive integer")
	v.Check(model.ExamType != nil, "exam_type", "must be provided")
	v.Check(len(model.ExamType) >= 1, "exam_type", "must contain at least 1 genre")
	v.Check(len(model.ExamType) <= 5, "exam_type", "must not contain more than 5 genres")
	v.Check(validator.Unique(model.ExamType), "exam_type", "must not contain duplicate values")
}

type ModelModel struct {
	DB *sql.DB
}

func (m ModelModel) Insert(model *ModelInfo) error {
	query := `
		INSERT INTO module_info (created_at, updated_at, module_name, module_duration, exam_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version`
	args := []any{model.CreatedAt, model.UpdatedAt, model.ModuleName, model.ModuleDuration, pq.Array(model.ExamType)}
	return m.DB.QueryRow(query, args...).Scan(&model.ID, &model.CreatedAt, &model.Version)
}

func (m ModelModel) Get(id int64) (*ModelInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound

	}
	query := `
	SELECT id, created_at, module_name, module_duration, exam_type, version
	FROM module_info
	WHERE id = $1`

	var model ModelInfo

	err := m.DB.QueryRow(query, id).Scan(
		&model.ID,
		&model.CreatedAt,
		&model.ModuleName,
		&model.UpdatedAt,
		&model.ModuleDuration,
		pq.Array(&model.ExamType),
		&model.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &model, nil
}

func (m ModelModel) Update(model *ModelInfo) error {
	query := `
		UPDATE module_info
		SET module_name = $1, module_duration = $2, exam_type = $3, version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING version`
	args := []any{
		model.ModuleName,
		model.ModuleDuration,
		pq.Array(model.ExamType),
		model.ID,
	}
	err := m.DB.QueryRow(query, args...).Scan(&model.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil

}

func (m ModelModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM module_info
		WHERE id = $1`
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
