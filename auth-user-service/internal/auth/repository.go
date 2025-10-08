package auth

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	CreateUser(email, passwordHash string) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(userID int) (*User, error)
	SaveToken(userID int, token string, expiresAt time.Time) error
	GetUserByToken(token string) (*User, error)
	UserExists(email string) (bool, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) bool
}

// Остальной код остается таким же как в предыдущей версии...
type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

type User struct {
	ID           int
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func (r *repository) CreateUser(email, passwordHash string) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		email, passwordHash,
	).Scan(&id)
	return id, err
}

func (r *repository) GetUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByID(userID int) (*User, error) {
	var user User
	err := r.db.QueryRow(
		"SELECT id, email, password_hash, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) SaveToken(userID int, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO auth_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, expiresAt,
	)
	return err
}

func (r *repository) GetUserByToken(token string) (*User, error) {
	var user User
	err := r.db.QueryRow(
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

func (r *repository) UserExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		email,
	).Scan(&exists)
	return exists, err
}

func (r *repository) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (r *repository) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
