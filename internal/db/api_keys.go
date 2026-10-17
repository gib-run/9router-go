package db

import (
	"database/sql"
	"fmt"
	"strconv"

	"9router/proxy/internal/models"
)

// ApiKey is a type alias for models.APIKey with token limits and quotas.
type ApiKey = models.APIKey

// scanApiKey reads an API key from any scanner implementing Scan.
func scanApiKey(row interface{ Scan(dest ...any) error }) (*models.APIKey, error) {
	var k models.APIKey
	dest, apply := apiKeyScan(&k)
	if err := row.Scan(dest...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan api key: %w", err)
	}
	apply()
	return &k, nil
}

// IncrementKeyUsedTokens increments the used_tokens counter for a key by integer ID.
func (r *Repo) IncrementKeyUsedTokens(id int64, tokens int64) error {
	idStr := strconv.FormatInt(id, 10)
	_, err := r.db.Exec(
		`UPDATE apiKeys SET used_tokens = used_tokens + ? WHERE id = ? OR id = ?`,
		tokens, id, idStr,
	)
	if err != nil {
		return fmt.Errorf("increment key used tokens %d: %w", id, err)
	}
	return nil
}

// ResetKeyUsedTokens resets the used_tokens counter to 0 for a key by integer ID.
func (r *Repo) ResetKeyUsedTokens(id int64) error {
	idStr := strconv.FormatInt(id, 10)
	_, err := r.db.Exec(
		`UPDATE apiKeys SET used_tokens = 0 WHERE id = ? OR id = ?`,
		id, idStr,
	)
	if err != nil {
		return fmt.Errorf("reset key used tokens %d: %w", id, err)
	}
	return nil
}

// IncrementKeyUsedTokensByID increments used_tokens for a key identified by string ID.
func (r *Repo) IncrementKeyUsedTokensByID(id string, tokens int64) error {
	_, err := r.db.Exec(
		`UPDATE apiKeys SET used_tokens = used_tokens + ? WHERE id = ?`,
		tokens, id,
	)
	if err != nil {
		return fmt.Errorf("increment key used tokens %s: %w", id, err)
	}
	return nil
}

// ResetKeyUsedTokensByID resets used_tokens to 0 for a key identified by string ID.
func (r *Repo) ResetKeyUsedTokensByID(id string) error {
	_, err := r.db.Exec(
		`UPDATE apiKeys SET used_tokens = 0 WHERE id = ?`,
		id,
	)
	if err != nil {
		return fmt.Errorf("reset key used tokens %s: %w", id, err)
	}
	return nil
}

// SetApiKeyTokenLimit updates the token_limit column for an API key.
func (r *Repo) SetApiKeyTokenLimit(id string, tokenLimit *int64) error {
	var val any
	if tokenLimit != nil {
		val = *tokenLimit
	}
	res, err := r.db.Exec(`UPDATE apiKeys SET token_limit = ? WHERE id = ?`, val, id)
	if err != nil {
		return fmt.Errorf("set api key token limit %s: %w", id, err)
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return fmt.Errorf("set api key token limit %s: %w", id, errNoSuchApiKey)
	}
	return nil
}
