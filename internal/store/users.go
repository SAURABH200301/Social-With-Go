package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at"`
	IsActive  bool     `json:"is_active"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(plainText string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainText), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.text = &plainText
	p.hash = hash
	return nil
}
func (p *password) Compare(text string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UsersStorage struct {
	db *sql.DB
}

func (s *UsersStorage) Create(ctx context.Context, tx *sql.Tx, user *Users) error {
	query := `INSERT INTO users (username, email, password, created_at) 
			VALUES ($1, $2, $3, NOW()) RETURNING id, created_at`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	err := tx.QueryRowContext(ctx, query, user.Username, user.Email, user.Password.hash).Scan(&user.ID, &user.CreatedAt)
	return err
}
func (s *UsersStorage) updateUser(ctx context.Context, tx *sql.Tx, user *Users) error {
	query := `UPDATE users SET is_active = $1, username = $2, email = $3 WHERE id = $4`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, user.IsActive, user.Username, user.Email, user.ID)
	return err
}

func (s *UsersStorage) GetByID(ctx context.Context, id int64) (*Users, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE id = $1 AND is_active = true`
	row := s.db.QueryRowContext(ctx, query, id)

	var user Users
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return &user, nil
}

func (s *UsersStorage) CreateAndInvite(ctx context.Context, user *Users, token string, invitationExp time.Duration) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.Create(ctx, tx, user); err != nil {
			return err
		}
		// create the user invite
		err := s.createUserInvitation(ctx, tx, token, invitationExp, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersStorage) createUserInvitation(ctx context.Context, tx *sql.Tx, token string, invitationExp time.Duration, userID int64) error {
	query := `INSERT INTO user_invitations (user_id, token, expiry) ValUES ($1, $2, $3)`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, userID, token, time.Now().Add(invitationExp))
	return err
}

func (s *UsersStorage) Activate(ctx context.Context, token string) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		user, err := s.getUserByInvitationToken(ctx, tx, token)
		if err != nil {
			return err
		}
		user.IsActive = true
		if err := s.updateUser(ctx, tx, user); err != nil {
			return err
		}
		if err := s.DeleteInvitationByUserID(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersStorage) getUserByInvitationToken(ctx context.Context, tx *sql.Tx, token string) (*Users, error) {
	query := `SELECT u.id, u.username, u.email, u.created_at , u.is_active
			FROM users u
			JOIN user_invitations ui ON u.id = ui.user_id
			WHERE ui.token = $1 AND ui.expiry > NOW()`

	hashToken := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hashToken[:])
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	row := tx.QueryRowContext(ctx, query, hashedToken)

	var user Users
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with token %s not found", token)
		}
		return nil, err
	}
	return &user, nil
}

func (s *UsersStorage) DeleteInvitationByUserID(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, userID)
	return err
}

func (s *UsersStorage) DeleteByID(ctx context.Context, id int64) error {
	return withTx(ctx, s.db, func(tx *sql.Tx) error {
		if err := s.deleteUserByID(ctx, tx, id); err != nil {
			return err
		}
		if err := s.DeleteInvitationByUserID(ctx, tx, id); err != nil {
			return err
		}
		return nil
	})
}

func (s *UsersStorage) deleteUserByID(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	_, err := tx.ExecContext(ctx, query, id)
	return err
}

func (s *UsersStorage) GetByEmail(ctx context.Context, email string) (*Users, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE email = $1 AND is_active = true`
	row := s.db.QueryRowContext(ctx, query, email)

	var user Users
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with Email %d not found", email)
		}
		return nil, err
	}
	return &user, nil
}
