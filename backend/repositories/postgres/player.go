package postgres

import (
	"context"
	"database/sql"
	"errors"
	"ligain/backend/models"
	"time"

	"github.com/google/uuid"
)

type PostgresPlayerRepository struct {
	db *sql.DB
}

func NewPostgresPlayerRepository(db *sql.DB) *PostgresPlayerRepository {
	return &PostgresPlayerRepository{db: db}
}

func (r *PostgresPlayerRepository) GetPlayer(playerId string) (models.Player, error) {
	var player models.PlayerData
	err := r.db.QueryRow("SELECT id, name FROM player WHERE id = $1", playerId).Scan(&player.ID, &player.Name)
	if err != nil {
		return &models.PlayerData{}, err
	}
	return &player, nil
}

func (r *PostgresPlayerRepository) GetPlayers(gameId string) ([]models.Player, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT p.id, p.name 
		FROM player p
		JOIN bet b ON p.id = b.player_id
		WHERE b.game_id = $1
	`, gameId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.PlayerData
		if err := rows.Scan(&player.ID, &player.Name); err != nil {
			return nil, err
		}
		players = append(players, &player)
	}
	return players, nil
}

// Authentication methods
func (r *PostgresPlayerRepository) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	query := `
		INSERT INTO player (name, email, provider, provider_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		player.Name, player.Email, player.Provider, player.ProviderID,
		time.Now(), time.Now()).Scan(&player.ID)
	return err
}

func (r *PostgresPlayerRepository) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	var player models.PlayerData
	query := `
		SELECT id, name, email, provider, provider_id, created_at, updated_at,
			avatar_object_key, avatar_signed_url, avatar_signed_url_expires_at
		FROM player WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&player.ID, &player.Name, &player.Email, &player.Provider, &player.ProviderID,
		&player.CreatedAt, &player.UpdatedAt,
		&player.AvatarObjectKey, &player.AvatarSignedURL, &player.AvatarSignedURLExpiresAt)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PostgresPlayerRepository) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	var player models.PlayerData
	query := `
		SELECT id, name, email, provider, provider_id, created_at, updated_at,
			avatar_object_key, avatar_signed_url, avatar_signed_url_expires_at
		FROM player WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&player.ID, &player.Name, &player.Email, &player.Provider, &player.ProviderID,
		&player.CreatedAt, &player.UpdatedAt,
		&player.AvatarObjectKey, &player.AvatarSignedURL, &player.AvatarSignedURLExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

func (r *PostgresPlayerRepository) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	var player models.PlayerData
	query := `
		SELECT id, name, email, provider, provider_id, created_at, updated_at,
			avatar_object_key, avatar_signed_url, avatar_signed_url_expires_at
		FROM player WHERE provider = $1 AND provider_id = $2
	`
	err := r.db.QueryRowContext(ctx, query, provider, providerID).Scan(
		&player.ID, &player.Name, &player.Email, &player.Provider, &player.ProviderID,
		&player.CreatedAt, &player.UpdatedAt,
		&player.AvatarObjectKey, &player.AvatarSignedURL, &player.AvatarSignedURLExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

func (r *PostgresPlayerRepository) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	var player models.PlayerData
	query := `
		SELECT id, name, email, provider, provider_id, created_at, updated_at,
			avatar_object_key, avatar_signed_url, avatar_signed_url_expires_at
		FROM player WHERE name = $1
	`
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&player.ID, &player.Name, &player.Email, &player.Provider, &player.ProviderID,
		&player.CreatedAt, &player.UpdatedAt,
		&player.AvatarObjectKey, &player.AvatarSignedURL, &player.AvatarSignedURLExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

func (r *PostgresPlayerRepository) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	query := `
		UPDATE player 
		SET name = $2, email = $3, provider = $4, provider_id = $5, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		player.ID, player.Name, player.Email, player.Provider, player.ProviderID)
	return err
}

func (r *PostgresPlayerRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	if token.ID == "" {
		token.ID = uuid.New().String()
	}

	token.CreatedAt = time.Now()

	query := `
		INSERT INTO auth_tokens (id, player_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.PlayerID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

func (r *PostgresPlayerRepository) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	query := `
		SELECT id, player_id, token, expires_at, created_at
		FROM auth_tokens
		WHERE token = $1
	`

	var authToken models.AuthToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&authToken.ID,
		&authToken.PlayerID,
		&authToken.Token,
		&authToken.ExpiresAt,
		&authToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &authToken, nil
}

func (r *PostgresPlayerRepository) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	query := `
		UPDATE auth_tokens 
		SET player_id = $2, token = $3, expires_at = $4, created_at = $5
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID,
		token.PlayerID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)
	return err
}

func (r *PostgresPlayerRepository) DeleteAuthToken(ctx context.Context, token string) error {
	query := `DELETE FROM auth_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, token)
	return err
}

func (r *PostgresPlayerRepository) DeleteExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM auth_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *PostgresPlayerRepository) DeletePlayer(ctx context.Context, playerID string) error {
	query := `DELETE FROM player WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, playerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresPlayerRepository) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	query := `
		UPDATE player
		SET avatar_object_key = $2,
			avatar_signed_url = $3,
			avatar_signed_url_expires_at = $4,
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, playerID, objectKey, signedURL, expiresAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresPlayerRepository) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	query := `
		UPDATE player
		SET avatar_signed_url = $2,
			avatar_signed_url_expires_at = $3,
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, playerID, signedURL, expiresAt)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresPlayerRepository) ClearAvatar(ctx context.Context, playerID string) error {
	query := `
		UPDATE player
		SET avatar_object_key = NULL,
			avatar_signed_url = NULL,
			avatar_signed_url_expires_at = NULL,
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, playerID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
