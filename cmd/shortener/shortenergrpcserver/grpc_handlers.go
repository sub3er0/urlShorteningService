package shortenergrpcserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"github.com/sub3er0/urlShorteningService/internal/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log"
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
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md["authorization"]; len(values) > 0 {
		}
	}
	var buf bytes.Buffer
	if compressedReq, ok := req.([]byte); ok {
		reader, err := gzip.NewReader(bytes.NewReader(compressedReq))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error decompressing request body: %v", err)
		}
		defer reader.Close()

		if _, err := io.Copy(&buf, reader); err != nil {
			return nil, status.Errorf(codes.Internal, "error reading decompressed request body: %v", err)
		}

		// Set the decompressed request back
		req = buf.Bytes() // adjust this according to the actual request type
	}

	// Call the next handler
	resp, err = handler(ctx, req)
	if err != nil {
		return resp, err
	}

	// Compressing the response if needed
	if info.FullMethod == "/your.package.Service/YourMethod" { // adjust accordingly
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(resp.([]byte)); err != nil { // assuming response is a byte slice
			return nil, status.Errorf(codes.Internal, "error compressing response body: %v", err)
		}
		if err := gz.Close(); err != nil {
			return nil, status.Errorf(codes.Internal, "error closing gzip writer: %v", err)
		}

		// Set the compressed response back
		resp = buf.Bytes() // adjust this according to the actual response type
	}

	return resp, nil
}

type GRPCHandlers struct {
	UnimplementedURLShortenerServer
	us *shortener.URLShortener
}

func NewGRPCHandlers(urlShortener *shortener.URLShortener) *GRPCHandlers {
	return &GRPCHandlers{us: urlShortener}
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
