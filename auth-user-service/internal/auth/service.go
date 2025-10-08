package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Register(email, password string) (string, error)
	Login(email, password string) (string, error)
	ValidateToken(token string) (*User, error)
}

type service struct {
	repo       Repository
	jwtManager *JWTManager
}

func NewService(repo Repository, jwtSecret string) Service {
	jwtManager := NewJWTManager(jwtSecret)
	return &service{
		repo:       repo,
		jwtManager: jwtManager,
	}
}

func (s *service) Register(email, password string) (string, error) {
	// Проверяем, существует ли пользователь
	exists, err := s.repo.UserExists(email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("user already exists")
	}

	// Создаем пользователя
	userID, err := s.repo.CreateUser(email, password)
	if err != nil {
		return "", err
	}

	// Генерируем JWT токен
	token, err := s.jwtManager.GenerateToken(userID, email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) Login(email, password string) (string, error) {
	// Получаем пользователя
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Генерируем JWT токен
	token, err := s.jwtManager.GenerateToken(user.ID, email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) ValidateToken(token string) (*User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Получаем пользователя из БД для проверки
	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
