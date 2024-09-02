package service

import (
	"context"
	"fmt"
	"time"

	"github.com/DarRo9/Test-task-BackDev/internal/config"
)

type Service struct {
	config             *config.Config
	storage            Storage
	tokenAuthenticator TokenAuthenticator
}

type TokenAuthenticator interface {
	CreateObjectJWT(userId string, ttl time.Duration) (string, error)
	CreateObjectRefreshToken() (string, error)
	HashToken(token string) ([]byte, error)
	CompareTokens(providedToken string, hashedToken []byte) bool
}

type Storage interface {
	InsertToken(ctx context.Context, userName string, refreshToken string, timeNow time.Time) error
	DeleteToken(ctx context.Context, refreshToken string) error
	DeleteTokensByUser(ctx context.Context, userName string) error
	SelectToken(ctx context.Context, oldRefreshToken string, CreateObjectRefreshToken string, userName string, timeNow time.Time) error
	CountTokens(ctx context.Context, userName string) (int64, error)
	GetTime(ctx context.Context, refreshToken string, userName string) (time.Time, error)
	GetTokenByUser(ctx context.Context, userName string) (string, error)
}

func CreateObject(config *config.Config, storage Storage, tokenAuthenticator TokenAuthenticator) (*Service, error) {
	return &Service{
		config:             config,
		storage:            storage,
		tokenAuthenticator: tokenAuthenticator}, nil
}

func (s *Service) MakeAccessToken(userName string) (string, error) {
	const op = "service.MakeAccessToken"

	accessToken, err := s.tokenAuthenticator.CreateObjectJWT(userName, s.config.AccessTokenTTL)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, nil
}

func (s *Service) RefreshToken(userName string) (string, error) {
	const op = "service.RefreshToken"

	refreshToken, err := s.tokenAuthenticator.CreateObjectRefreshToken()
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshToken, nil
}

func (s *Service) TakeValidToken(tokenFromHeader string, userName string) bool {
	tokenFromDB, err := s.getTokenFromDB(userName)
	if err != nil {
		return false
	}

	if ok := s.tokenAuthenticator.CompareTokens(tokenFromHeader, []byte(tokenFromDB)); !ok {
		return false
	}

	if ok, err := s.checkTokenTtl(tokenFromDB, userName, time.Now()); err != nil || !ok {
		return false
	}

	return true
}

func (s *Service) checkTokenTtl(tokenFromDB string, userName string, time time.Time) (bool, error) {
	const op = "service.checkTokenTtl"

	createdTime, err := s.storage.GetTime(context.TODO(), tokenFromDB, userName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if createdTime.Add(s.config.JWT.RefreshTokenTTL).Before(time) {
		if err := s.storage.DeleteToken(context.TODO(), tokenFromDB); err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}

		return false, nil
	}

	return true, nil
}

func (s *Service) getTokenFromDB(userName string) (string, error) {
	const op = "service.getTokenFromDB"

	refreshTokenFromDB, err := s.storage.GetTokenByUser(context.Background(), userName)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return refreshTokenFromDB, nil
}

func (s *Service) SelectToken(CreateObjectToken string, userName string) error {
	const op = "service.switchToken"

	oldToken, err := s.getTokenFromDB(userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	hashedToken, err := s.tokenAuthenticator.HashToken(CreateObjectToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.storage.SelectToken(context.TODO(), oldToken, string(hashedToken), userName, time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) CheckCountTokens(userName string) error {
	const op = "service.CheckCountTokens"

	count, err := s.storage.CountTokens(context.TODO(), userName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if count > 0 {
		if err := s.storage.DeleteTokensByUser(context.TODO(), userName); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *Service) InsertToken(refreshToken string, userName string) error {
	const op = "service.InsertToken"

	hashedToken, err := s.tokenAuthenticator.HashToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.storage.InsertToken(context.TODO(), userName, string(hashedToken), time.Now()); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
