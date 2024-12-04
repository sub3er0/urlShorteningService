package grpc_server

import (
	"context"
	"errors"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/repository"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"log"
)

type GRPCHandlers struct {
	UnimplementedURLShortenerServer
	userRepository *repository.UserRepository
	urlRepository  *repository.URLRepository
	CookieManager  cookie.CookieManagerInterface
}

func NewGRPCHandlers(urlRepository *repository.URLRepository, userRepository *repository.UserRepository, cookieManager cookie.CookieManager) *GRPCHandlers {
	return &GRPCHandlers{
		urlRepository:  urlRepository,
		userRepository: userRepository,
		CookieManager:  &cookieManager,
	}
}

// ShortenURL обрабатывает gRPC-запрос для сокращения URL.
func (h *GRPCHandlers) ShortenURL(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	shortKey, err := h.urlRepository.GetShortURL(req.GetOriginalUrl())

	if err == nil {
		return &ShortenResponse{ShortenedUrl: shortKey}, nil
	}

	shortKey = shortener.GenerateShortKey()
	err = h.urlRepository.Save(shortKey, req.GetOriginalUrl(), h.CookieManager.GetActualCookieValue())

	if err != nil {
		return &ShortenResponse{ShortenedUrl: ""}, err
	}

	return &ShortenResponse{ShortenedUrl: shortKey}, nil
}

// GetURL обрабатывает gRPC-запрос для получения оригинального URL.
func (h *GRPCHandlers) GetURL(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	log.Printf(req.GetShortenedUrl())
	originalURL, ok := h.urlRepository.GetURL(req.GetShortenedUrl())
	if ok != true {
		return nil, errors.New("NotFound")
	}
	return &GetResponse{OriginalUrl: originalURL.URL}, nil
}

// GetInternalStats обрабатывает gRPC-запрос для получения статистики.
func (h *GRPCHandlers) GetInternalStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	urlsCount, err := h.urlRepository.GetURLCount()
	if err != nil {
		return nil, err
	}

	usersCount, err := h.userRepository.GetUsersCount()
	if err != nil {
		return nil, err
	}

	return &StatsResponse{Urls: int64(urlsCount), Users: int64(usersCount)}, nil
}

// Ping обрабатывает gRPC-запрос для проверки доступности.
func (h *GRPCHandlers) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{Message: "pong"}, nil
}
