package auth

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) UserExists(email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := p.db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

func (p *PostgresRepository) CreateUser(email, passwordHash string) (int, error) {
	query := "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id"
	var userID int
	err := p.db.QueryRow(query, email, passwordHash).Scan(&userID)
	return userID, err
}

func (p *PostgresRepository) GetUserByEmail(email string) (*User, error) {
	query := "SELECT id, email, password_hash, created_at FROM users WHERE email = $1"
	user := &User{}
	err := p.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (p *PostgresRepository) GetUserByID(userID int) (*User, error) {
	query := "SELECT id, email, password_hash, created_at FROM users WHERE id = $1"
	user := &User{}
	err := p.db.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (p *PostgresRepository) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (p *PostgresRepository) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Сохранение токена (для обратной совместимости)
func (p *PostgresRepository) SaveToken(userID int, token string, expiresAt time.Time) error {
	_, err := p.db.Exec(
		"INSERT INTO auth_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	return err
}

// Получение пользователя по токену (для обратной совместимости)
func (p *PostgresRepository) GetUserByToken(token string) (*User, error) {
	var user User
	err := p.db.QueryRow(
		`SELECT u.id, u.email, u.password_hash, u.created_at 
		 FROM users u 
		 JOIN auth_tokens t ON u.id = t.user_id 
		 WHERE t.token = $1 AND t.expires_at > NOW()`,
		token,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
