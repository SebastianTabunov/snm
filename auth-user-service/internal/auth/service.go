package auth

import (
	"errors"
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

	// Хеширование пароля
	passwordHash, err := s.repo.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Создание пользователя
	userID, err := s.repo.CreateUser(email, passwordHash)
	if err != nil {
		return "", err
	}

	// Генерация JWT токена
	token, err := s.jwtManager.GenerateToken(userID, email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) Login(email, password string) (string, error) {
	// Получение пользователя
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Проверка пароля
	if !s.repo.VerifyPassword(user.PasswordHash, password) {
		return "", errors.New("invalid credentials")
	}

	// Генерация JWT токена
	token, err := s.jwtManager.GenerateToken(user.ID, email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *service) ValidateToken(token string) (*User, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Получаем пользователя из БД для проверки
	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
