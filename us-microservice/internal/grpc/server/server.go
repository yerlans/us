package server

import (
	"context"
	"errors"
	pb "github.com/yerlans/us-protos/gen/us-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"urlSh/internal/storage"
)

type URLShortener interface {
	ShortenURL(ctx context.Context, originalURL string) (shortURL string, err error)
	GetOriginalURL(ctx context.Context, shortURL string) (originalURL string, err error)
}

type serverAPI struct {
	pb.UnimplementedUrlShorteningServiceServer
	shortener URLShortener
}

func Register(gRPCServer *grpc.Server, shortener URLShortener) {
	pb.RegisterUrlShorteningServiceServer(gRPCServer, &serverAPI{shortener: shortener})
}

func (s *serverAPI) ShortenUrl(
	ctx context.Context,
	in *pb.ShortenUrlRequest,
) (*pb.ShortenUrlResponse, error) {
	if in.OriginalUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "original_url is required")
	}

	shortURL, err := s.shortener.ShortenURL(ctx, in.GetOriginalUrl())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	return &pb.ShortenUrlResponse{ShortUrl: shortURL}, nil
}

func (s *serverAPI) GetOriginalUrl(
	ctx context.Context,
	in *pb.GetOriginalUrlRequest,
) (*pb.GetOriginalUrlResponse, error) {
	if in.ShortUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "short_url is required")
	}

	originalURL, err := s.shortener.GetOriginalURL(ctx, in.GetShortUrl())
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return nil, status.Error(codes.NotFound, "short URL not found")
		}
		return nil, status.Error(codes.Internal, "failed to get original URL")
	}

	return &pb.GetOriginalUrlResponse{OriginalUrl: originalURL}, nil
}
