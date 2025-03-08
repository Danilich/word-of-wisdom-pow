package services

import (
	"context"
	"fmt"
)

type QuotesService struct {
	repository QuotesRepository
}

func NewQuotesService(repository QuotesRepository) *QuotesService {
	return &QuotesService{
		repository: repository,
	}
}

// GetRandomQuote returns a random quote
func (s *QuotesService) GetRandomQuote(ctx context.Context) (string, error) {
	quote, err := s.repository.GetRandom(ctx)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s - %s", quote.Text, quote.Author), nil
}
