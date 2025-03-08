package tcpserver

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
)

// HandlerQuote Command constants
const (
	HandlerQuote = iota + 1
)

// QuoteHandler creates a handler for quotes
func QuoteHandler(quotesService QuotesService) HandlerFunc {
	return func(ctx context.Context, conn net.Conn) error {
		remoteAddr := conn.RemoteAddr().String()
		quote, err := quotesService.GetRandomQuote(ctx)

		if err != nil {
			if _, writeErr := conn.Write([]byte(QuoteServerError)); writeErr != nil {
				return fmt.Errorf("failed to write error message: %w (original error: %v)", writeErr, err)
			}
			return fmt.Errorf("failed to get quote: %w", err)
		}

		if _, err := conn.Write([]byte(fmt.Sprintf(ResponseGeneratedWisdom, quote))); err != nil {
			return fmt.Errorf("failed to send quote: %w", err)
		}

		log.Info().Str("client_addr", remoteAddr).Msg("Quote sent")

		return nil
	}
}
