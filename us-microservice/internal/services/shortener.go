package services

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

type UrlStorage interface {
	SaveURL(ctx context.Context, urlToSave, alias string) (string, error)
	GetURL(ctx context.Context, alias string) (string, error)
}

type CacheStorage interface {
	SaveURL(ctx context.Context, originalURL string, alias string, expiration time.Duration) error
	GetURL(ctx context.Context, alias string) (string, error)
}

type URLShortener struct {
	log     *slog.Logger
	storage UrlStorage
	cache   CacheStorage
	ttl     time.Duration
}

func New(log *slog.Logger,
	storage UrlStorage,
	cache CacheStorage,
	ttl time.Duration) *URLShortener {
	return &URLShortener{
		log:     log,
		storage: storage,
		cache:   cache,
		ttl:     ttl,
	}
}

func (u *URLShortener) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	//TODO: check if url already exists, not it checks (url, alias) in db, but alias is random everytime
	u.log.Info("attempting to shorten URL")
	alias := generateShortURL(5)
	url, err := u.storage.SaveURL(ctx, originalURL, alias)
	if err != nil {
		return "", err
	}
	err = u.cache.SaveURL(ctx, originalURL, alias, u.ttl)
	if err != nil {
		return "", err
	}
	return url, nil
}

// GetOriginalURL retrieves the original URL for a given short URL.
func (u *URLShortener) GetOriginalURL(
	ctx context.Context,
	shortURL string,
) (string, error) {

	u.log.Info("attempting to fetch original URL")
	getURL, err := u.cache.GetURL(ctx, shortURL)
	if err == nil && getURL != "" {
		return getURL, nil
	}
	url, err := u.storage.GetURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	return url, nil
}

// Helper function to generate short URL
func generateShortURL(size int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}
