package services

import (
	"context"
	"word-of-wisdom-pow/internal/server/domain"
)

// QuotesRepository is for accessing quotes
type QuotesRepository interface {
	GetRandom(ctx context.Context) (domain.Quote, error)
}
