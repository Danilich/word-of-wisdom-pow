package tcpserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog/log"

	"word-of-wisdom-pow/internal/server/config"
	"word-of-wisdom-pow/internal/server/services"
)

// SolutionSize represents the size of the PoW solution in bytes
const SolutionSize = 8
const clientAddrKey = "client_addr"

// PowHandler represents handler for client connections with PoW
type PowHandler struct {
	powService *services.PowService
	config     config.Config
	ctx        context.Context
	router     *Router
}

func NewConnectionHandler(ctx context.Context, powService *services.PowService, cfg config.Config, r *Router) *PowHandler {
	return &PowHandler{
		powService: powService,
		config:     cfg,
		ctx:        ctx,
		router:     r,
	}
}

// HandleClient processes a client connection
func (h *PowHandler) HandleClient(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	log.Info().Str(clientAddrKey, remoteAddr).Msg("New connection established")

	connCtx, cancel := context.WithTimeout(h.ctx, h.config.ConnectionTimeout)

	defer cancel()

	if err := conn.SetDeadline(time.Now().Add(h.config.ConnectionTimeout)); err != nil {
		log.Error().Err(err).Str(clientAddrKey, remoteAddr).Msg("Failed to set connection deadline")
		return
	}

	if err := h.processPowChallenge(connCtx, conn); err != nil {
		log.Error().Err(err).Str(clientAddrKey, remoteAddr).Msg("Failed to process PoW challenge")
		return
	}

	cmdID, err := h.readCommand(conn)
	if err != nil {
		log.Error().Err(err).Str(clientAddrKey, remoteAddr).Msg("Failed to read command")
		return
	}

	log.Debug().Int("command", cmdID).Str(clientAddrKey, remoteAddr).Msg("Received command")

	if err := h.router.HandleCommand(connCtx, cmdID, conn); err != nil {
		log.Error().Err(err).Str(clientAddrKey, remoteAddr).Int("command", cmdID).Msg("Failed to handle command")
		return
	}
}

// readCommand reads a single byte command from the connection
func (h *PowHandler) readCommand(conn net.Conn) (int, error) {
	cmdByte := make([]byte, 1)

	if _, err := io.ReadFull(conn, cmdByte); err != nil {
		var responseMsg string
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			responseMsg = ResponseCommandTimeout
			log.Warn().Str("client_addr", conn.RemoteAddr().String()).Msg("Timeout reading command")
		} else {
			responseMsg = ResponseFailedReadCommand
		}

		if writeErr := h.sendErrorResponse(conn, responseMsg); writeErr != nil {
			log.Error().Err(writeErr).Str("client_addr", conn.RemoteAddr().String()).Msg("Failed to send error response")
		}
		return 0, fmt.Errorf("failed to read command byte: %w", err)
	}

	return int(cmdByte[0]), nil
}

func (h *PowHandler) processPowChallenge(ctx context.Context, conn net.Conn) error {
	// Generate and send challenge
	challenge := h.powService.GenerateChallenge()

	if err := h.sendChallenge(conn, challenge); err != nil {
		return err
	}

	// Read solution with timeout
	solution, err := h.readSolutionWithTimeout(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to read solution: %w", err)
	}

	// Verify solution
	if !h.powService.VerifyProof(challenge, solution) {
		if err := h.sendErrorResponse(conn, ResponseInvalidPowSolution); err != nil {
			return fmt.Errorf("failed to send invalid solution response: %w", err)
		}
		return fmt.Errorf("invalid solution provided")
	}

	// Send success response
	if err := h.sendSuccessResponse(conn); err != nil {
		return err
	}

	return nil
}

// readSolutionWithTimeout reads the PoW solution from the client with a timeout
func (h *PowHandler) readSolutionWithTimeout(ctx context.Context, conn net.Conn) ([]byte, error) {
	type result struct {
		solution []byte
		err      error
	}
	resultCh := make(chan result, 1)

	go func() {
		solution := make([]byte, SolutionSize)
		_, err := io.ReadFull(conn, solution)
		select {
		case <-ctx.Done():
		case resultCh <- result{solution, err}:
		}
	}()

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout reading solution: %w", ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			return nil, fmt.Errorf("failed to read solution: %w", res.err)
		}
		return res.solution, nil
	}
}

// sendChallenge sends the PoW challenge to the client
func (h *PowHandler) sendChallenge(conn net.Conn, challenge []byte) error {
	combinedData := make([]byte, len(challenge)+1)
	combinedData[0] = h.config.Difficulty
	copy(combinedData[1:], challenge)

	if _, err := conn.Write(combinedData); err != nil {
		return fmt.Errorf("failed to send challenge with difficulty: %w", err)
	}

	log.Debug().Bytes("challenge", challenge).Uint8("difficulty", h.config.Difficulty).Msg("Sent PoW challenge")
	return nil
}

// sendSuccessResponse sends a success message to the client
func (h *PowHandler) sendSuccessResponse(conn net.Conn) error {
	_, err := conn.Write([]byte(ResponsePowVerificationSuccess))

	if err != nil {
		return fmt.Errorf("failed to send success response: %w", err)
	}
	return nil
}

// sendErrorResponse sends an error message to the client
func (h *PowHandler) sendErrorResponse(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}
