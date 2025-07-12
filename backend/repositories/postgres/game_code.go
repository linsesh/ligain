package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresGameCodeRepository struct {
	db *sql.DB
}

func NewPostgresGameCodeRepository(db *sql.DB) repositories.GameCodeRepository {
	return &PostgresGameCodeRepository{db: db}
}

func (r *PostgresGameCodeRepository) CreateGameCode(gameCode *models.GameCode) error {
	query := `
		INSERT INTO game_codes (game_id, code, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRow(
		query,
		gameCode.GameID,
		gameCode.Code,
		gameCode.ExpiresAt,
	).Scan(&gameCode.ID, &gameCode.CreatedAt)

	if err != nil {
		return fmt.Errorf("error creating game code: %v", err)
	}

	return nil
}

func (r *PostgresGameCodeRepository) GetGameCodeByCode(code string) (*models.GameCode, error) {
	query := `
		SELECT id, game_id, code, created_at, expires_at
		FROM game_codes
		WHERE code = $1 AND expires_at > NOW()`

	var gameCode models.GameCode
	err := r.db.QueryRow(query, code).Scan(
		&gameCode.ID,
		&gameCode.GameID,
		&gameCode.Code,
		&gameCode.CreatedAt,
		&gameCode.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, repositories.ErrGameCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting game code: %v", err)
	}

	return &gameCode, nil
}

func (r *PostgresGameCodeRepository) GetGameCodeByGameID(gameID string) (*models.GameCode, error) {
	query := `
		SELECT id, game_id, code, created_at, expires_at
		FROM game_codes
		WHERE game_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`

	var gameCode models.GameCode
	err := r.db.QueryRow(query, gameID).Scan(
		&gameCode.ID,
		&gameCode.GameID,
		&gameCode.Code,
		&gameCode.CreatedAt,
		&gameCode.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, repositories.ErrGameCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error getting game code by game ID: %v", err)
	}

	return &gameCode, nil
}

func (r *PostgresGameCodeRepository) DeleteExpiredCodes() error {
	query := `DELETE FROM game_codes WHERE expires_at <= NOW()`

	result, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("error deleting expired codes: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Deleted %d expired game codes\n", rowsAffected)
	}

	return nil
}

func (r *PostgresGameCodeRepository) CodeExists(code string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM game_codes WHERE code = $1 AND expires_at > NOW())`

	var exists bool
	err := r.db.QueryRow(query, code).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if code exists: %v", err)
	}

	return exists, nil
}

func (r *PostgresGameCodeRepository) DeleteGameCode(code string) error {
	query := `DELETE FROM game_codes WHERE code = $1`

	result, err := r.db.Exec(query, code)
	if err != nil {
		return fmt.Errorf("error deleting game code: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repositories.ErrGameCodeNotFound
	}

	return nil
}
