package shortenergrpcserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"github.com/sub3er0/urlShorteningService/internal/cookie"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"github.com/sub3er0/urlShorteningService/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strings"
	"time"
)

// LoggingInterceptor - логирующий интерсептор
func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	method := info.FullMethod
	resp, err = handler(ctx, req)

	duration := time.Since(start)

	log.Printf("Request: method=%s, duration=%v, error=%v", method, duration, err)
	return resp, err
}

// GzipInterceptor - gzip интерсептор
func GzipInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, status.Errorf(codes.Internal, "missing metadata")
	}

	var buf bytes.Buffer
	if accepts := md.Get("Accept-Encoding"); len(accepts) > 0 && strings.Contains(accepts[0], "gzip") {
		if compressedReq, ok := req.([]byte); ok {
			reader, err := gzip.NewReader(bytes.NewReader(compressedReq))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "error decompressing request body: %v", err)
			}
			defer reader.Close()

			if _, err := io.Copy(&buf, reader); err != nil {
				return nil, status.Errorf(codes.Internal, "error reading decompressed request body: %v", err)
			}

			req = buf.Bytes()
		}
	}

	resp, err = handler(ctx, req)
	if err != nil {
		return resp, err
	}

	if accepts := md.Get("Content-Encoding"); len(accepts) > 0 && strings.Contains(accepts[0], "gzip") {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(resp.([]byte)); err != nil { // assuming response is a byte slice
			return nil, status.Errorf(codes.Internal, "error compressing response body: %v", err)
		}
		if err := gz.Close(); err != nil {
			return nil, status.Errorf(codes.Internal, "error closing gzip writer: %v", err)
		}

		resp = buf.Bytes()
	}

	return resp, nil
}

// CookieAuthInterceptor Интерсептор для авторизации
func CookieAuthInterceptor(cookieManager *cookie.CookieManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if info.FullMethod == "/grpc_server.URLShortener/GetUserUrls" {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
			}

			// Извлекаем значение куки
			cookies := md[cookie.CookieName]
			if len(cookies) == 0 || !cookie.VerifyCookie(cookies[0]) {
				return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
			}

			userID, ok := cookie.GetUserIDFromCookie(cookies[0])
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
			}

			isUserExist := cookieManager.Storage.IsUserExist(userID)
			if !isUserExist {
				return nil, status.Errorf(codes.Unauthenticated, "Unauthorized")
			}
		}

		return handler(ctx, req)
	}
}

// CookieInterceptor Интерсептор для проверки куки
func CookieInterceptor(cookieManager *cookie.CookieManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		var userID string
		createNewCookie := false

		cookieValues := md[cookie.CookieName] // Извлекаем куку из метаданных
		if len(cookieValues) == 0 || !cookie.VerifyCookie(cookieValues[0]) {
			createNewCookie = true
		} else {
			userID, _ = cookie.GetUserIDFromCookie(cookieValues[0])
			isUserExist := cookieManager.Storage.IsUserExist(userID)

			if !isUserExist {
				createNewCookie = true
			}
		}

		if createNewCookie {
			userID = cookie.GenerateUserID()
			newCookieValue := userID + "." + cookie.SignCookie(userID)
			cookieManager.Storage.SaveUser(userID)

			// Задаем новую куку в метаданные
			newMD := metadata.Pairs(cookie.CookieName, newCookieValue)
			grpc.SetHeader(ctx, newMD)

			cookieManager.ActualCookieValue = userID
		}

		return handler(ctx, req)
	}
}

type GRPCHandlers struct {
	UnimplementedURLShortenerServer
	us *shortener.URLShortener
	cm *cookie.CookieManager
}

func NewGRPCHandlers(urlShortener *shortener.URLShortener, cookieManager *cookie.CookieManager) *GRPCHandlers {
	return &GRPCHandlers{us: urlShortener, cm: cookieManager}
}

// DeleteUserUrls удаляет короткие URL
func (h *GRPCHandlers) DeleteUserUrls(ctx context.Context, req *DeleteUserUrlsRequest) (*DeleteUserUrlsResponse, error) {
	for _, shortURL := range req.GetShortUrls() {
		h.us.RemoveChan <- shortURL
	}

	return &DeleteUserUrlsResponse{Message: "URLs deleted successfully"}, nil
}

// GetUserUrls получает URL пользователя
func (h *GRPCHandlers) GetUserUrls(ctx context.Context, req *UserRequest) (*UserUrlsResponse, error) {
	userID := h.us.CookieManager.GetActualCookieValue()
	if userID == "" {
		return nil, status.Errorf(codes.Unauthenticated, "User not authenticated")
	}

	urls, err := h.us.UserRepository.GetUserUrls(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal Server Error")
	}

	responseUrls := make([]*UserUrlsResponseBodyItem, 0, len(urls))

	for _, url := range urls {
		responseUrls = append(responseUrls, &UserUrlsResponseBodyItem{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}

	return &UserUrlsResponse{Urls: responseUrls}, nil
}

// BatchShortenURL обрабатывает пакетные запросы на сокращение URL
func (h *GRPCHandlers) BatchShortenURL(ctx context.Context, req *BatchShortenRequest) (*BatchShortenResponse, error) {
	var responseBodyBatch []*BatchResponseBodyItem
	var dataStorageRows []storage.DataStorageRow

	for _, requestBodyRow := range req.GetOriginalUrls() {
		shortKey, getShortURLError := h.us.GetShortURL(requestBodyRow.OriginalUrl)
		if getShortURLError != nil {
			shortKey = shortener.GenerateShortKey()
		}

		responseBody := &BatchResponseBodyItem{
			CorrelationId: requestBodyRow.CorrelationId,
			ShortUrl:      h.us.BaseURL + shortKey,
		}

		responseBodyBatch = append(responseBodyBatch, responseBody)

		dataStorageRow := storage.DataStorageRow{
			ShortURL: shortKey,
			URL:      requestBodyRow.GetOriginalUrl(),
			UserID:   h.us.CookieManager.GetActualCookieValue(),
		}
		dataStorageRows = append(dataStorageRows, dataStorageRow)

		if len(dataStorageRows) == 1000 {
			err := h.us.URLRepository.SaveBatch(dataStorageRows)
			log.Printf("ERROR = %v", err)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Internal Server Error")
			}
			dataStorageRows = dataStorageRows[:0]
		}
	}

	if len(dataStorageRows) > 0 {
		err := h.us.URLRepository.SaveBatch(dataStorageRows)
		log.Printf("ERROR = %v", err)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal Server Error")
		}
	}

	return &BatchShortenResponse{ShortUrls: responseBodyBatch}, nil
}

// ShortenURL обрабатывает gRPC-запрос для сокращения URL.
func (h *GRPCHandlers) ShortenURL(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md["authorization"]; len(values) > 0 {
		}
	}
	shortKey, err := h.us.GetShortKey(req.GetOriginalUrl())

	if errors.Is(err, shortener.ErrShortURLExists) {
		return &ShortenResponse{ShortenedUrl: shortKey, Message: "Already exist"}, nil
	}

	shortKey = shortener.GenerateShortKey()
	err = h.us.URLRepository.Save(shortKey, req.GetOriginalUrl(), h.us.CookieManager.GetActualCookieValue())

	if err != nil {
		return &ShortenResponse{ShortenedUrl: ""}, err
	}

	return &ShortenResponse{ShortenedUrl: shortKey}, nil
}

// GetURL обрабатывает gRPC-запрос для получения оригинального URL.
func (h *GRPCHandlers) GetURL(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	originalURL, ok := h.us.URLRepository.GetURL(req.GetShortenedUrl())
	if !ok {
		return nil, errors.New("NotFound")
	}
	return &GetResponse{OriginalUrl: originalURL.URL}, nil
}

// GetInternalStats обрабатывает gRPC-запрос для получения статистики.
func (h *GRPCHandlers) GetInternalStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	urlsCount, err := h.us.URLRepository.GetURLCount()
	if err != nil {
		return nil, err
	}

	usersCount, err := h.us.UserRepository.GetUsersCount()
	if err != nil {
		return nil, err
	}

	return &StatsResponse{Urls: int64(urlsCount), Users: int64(usersCount)}, nil
}

// Ping обрабатывает gRPC-запрос для проверки доступности.
func (h *GRPCHandlers) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	return &PingResponse{Message: "pong"}, nil
}
