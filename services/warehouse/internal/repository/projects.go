package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/radonezhsklad/warehouse/internal/models"
)

type ProjectRepo struct{ db *pgxpool.Pool }

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo { return &ProjectRepo{db: db} }

func (r *ProjectRepo) Create(ctx context.Context, p *models.Project) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO projects (external_id, name, archived)
 VALUES (COALESCE($1, gen_random_uuid()), $2, FALSE)
 RETURNING id, external_id, name, archived, created_at, updated_at`,
		p.ExternalID, p.Name,
	).Scan(&p.ID, &p.ExternalID, &p.Name, &p.Archived, &p.CreatedAt, &p.UpdatedAt)
}

func (r *ProjectRepo) Get(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := r.db.QueryRow(ctx,
		`SELECT id, external_id, name, archived, created_at, updated_at
 FROM projects WHERE id = $1`, id,
	).Scan(&p.ID, &p.ExternalID, &p.Name, &p.Archived, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProjectRepo) GetByExternalID(ctx context.Context, ext uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := r.db.QueryRow(ctx,
		`SELECT id, external_id, name, archived, created_at, updated_at
 FROM projects WHERE external_id = $1`, ext,
	).Scan(&p.ID, &p.ExternalID, &p.Name, &p.Archived, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}

func (r *ProjectRepo) List(ctx context.Context, includeArchived bool) ([]models.Project, error) {
	q := `SELECT id, external_id, name, archived, created_at, updated_at FROM projects`
	if !includeArchived {
		q += ` WHERE archived = FALSE`
	}
	q += ` ORDER BY name`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Project{}
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.ExternalID, &p.Name, &p.Archived, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectRepo) Update(ctx context.Context, p *models.Project) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET name=$2, archived=$3, updated_at=NOW() WHERE id=$1`,
		p.ID, p.Name, p.Archived,
	)
	return err
}

func (r *ProjectRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
