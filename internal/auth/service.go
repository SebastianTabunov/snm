package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(email, password, firstName, lastName string) (*User, error)
	Login(email, password string) (*User, error)
	GenerateToken(userID int, email string) (string, error)
	ValidateToken(tokenString string) (int, string, error)
	GetUserByID(userID int) (*User, error)
}

type service struct {
	repo      Repository
	jwtSecret string
}

func NewService(repo Repository, jwtSecret string) Service {
	return &service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(email, password, firstName, lastName string) (*User, error) {
	// Проверяем существует ли пользователь через UserExists
	exists, err := s.repo.UserExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user already exists")
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создаем пользователя через репозиторий
	userID, err := s.repo.CreateUser(email, string(hashedPassword), firstName, lastName)
	if err != nil {
		return nil, err
	}

	// Получаем созданного пользователя
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Login(email, password string) (*User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *service) GenerateToken(userID int, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) ValidateToken(tokenString string) (int, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int(claims["user_id"].(float64))
		email := claims["email"].(string)
		return userID, email, nil
	}

	return 0, "", errors.New("invalid token")
}

func (s *service) GetUserByID(userID int) (*User, error) {
	return s.repo.GetUserByID(userID)
}
