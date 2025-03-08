package tcpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mocks
type MockQuotesService struct {
	mock.Mock
}

func (m *MockQuotesService) GetRandomQuote(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

type MockConn struct {
	mock.Mock
	writtenData []byte
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	m.writtenData = append(m.writtenData, b...)
	args := m.Called(b)
	return args.Int(0), args.Error(1)
}

func (m *MockConn) RemoteAddr() net.Addr { return &mockAddr{"127.0.0.1:1234"} }

func (m *MockConn) Read(b []byte) (int, error)         { return 0, nil }
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

type mockAddr struct{ addr string }

func (m *mockAddr) Network() string { return "tcp" }
func (m *mockAddr) String() string  { return m.addr }

func TestQuoteHandler(t *testing.T) {
	tests := []struct {
		name     string
		quote    string
		quoteErr error
		writeErr error
		wantErr  string
		ctx      context.Context
	}{
		{
			name:  "Success",
			quote: "Test quote",
			ctx:   context.Background(),
		},
		{
			name:  "Empty Quote",
			quote: "",
			ctx:   context.Background(),
		},
		{
			name:     "Service Error",
			quoteErr: errors.New("service error"),
			wantErr:  "failed to get quote",
			ctx:      context.Background(),
		},
		{
			name:     "Write Error",
			quote:    "Test quote",
			writeErr: errors.New("write error"),
			wantErr:  "failed to send quote",
			ctx:      context.Background(),
		},
		{
			name:     "Canceled Context",
			quoteErr: context.Canceled,
			wantErr:  "failed to get quote",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockQuotesService)
			conn := new(MockConn)

			svc.On("GetRandomQuote", mock.Anything).Return(tt.quote, tt.quoteErr)

			// Configure expected response
			if tt.quoteErr != nil {
				size := 0
				if tt.writeErr == nil {
					size = len(QuoteServerError)
				}
				conn.On("Write", []byte(QuoteServerError)).Return(size, tt.writeErr)
			} else {
				resp := fmt.Sprintf(ResponseGeneratedWisdom, tt.quote)
				size := 0
				if tt.writeErr == nil {
					size = len(resp)
				}
				conn.On("Write", []byte(resp)).Return(size, tt.writeErr)
			}

			// Execute
			err := QuoteHandler(svc)(tt.ctx, conn)

			// Verify
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				if tt.quote != "" {
					assert.Contains(t, string(conn.writtenData), tt.quote)
				} else {
					assert.Contains(t, string(conn.writtenData), "Generated wisdom: ")
				}
			}

			svc.AssertExpectations(t)
			conn.AssertExpectations(t)
		})
	}
}
