package user

import (
	"context"
	"fmt"
	"time"

	"auth-user-service/internal/redis"
)

type Service interface {
	GetProfile(userID int) (*Profile, error)
	UpdateProfile(userID int, profile *Profile) error
}

type service struct {
	repo  Repository
	redis *redis.Client
}

func NewService(repo Repository, redisClient *redis.Client) Service {
	return &service{
		repo:  repo,
		redis: redisClient,
	}
}

func (s *service) GetProfile(userID int) (*Profile, error) {
	// Пробуем получить из кэша Redis
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		var cachedProfile Profile

		ctx := context.Background()
		err := s.redis.Get(ctx, cacheKey, &cachedProfile)
		if err == nil {
			// Нашли в кэше
			return &cachedProfile, nil
		}
	}

	// Не нашли в кэше, получаем из БД
	profile, err := s.repo.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кэш
	if s.redis != nil && profile != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		ctx := context.Background()
		err := s.redis.Set(ctx, cacheKey, profile, 10*time.Minute)
		if err != nil {
			return nil, err
		} // Кэшируем на 10 минут
	}

	return profile, nil
}

func (s *service) UpdateProfile(userID int, profile *Profile) error {
	// Обновляем в БД
	err := s.repo.UpdateProfile(userID, profile)
	if err != nil {
		return err
	}

	// Инвалидируем кэш
	if s.redis != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		ctx := context.Background()
		err := s.redis.Delete(ctx, cacheKey)
		if err != nil {
			return err
		}
	}

	return nil
}
