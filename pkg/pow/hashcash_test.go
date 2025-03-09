package pow

import (
	"context"
	"encoding/binary"
	"testing"
)

func TestNewHashcash(t *testing.T) {
	tests := []struct {
		name       string
		difficulty uint8
		want       uint8
	}{
		{"valid difficulty", 3, 3},
		{"zero uses default", 0, DefaultDifficulty},
		{"max value", 255, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if h := NewHashcash(tt.difficulty); h.Difficulty != tt.want {
				t.Errorf("got difficulty %d, want %d", h.Difficulty, tt.want)
			}
		})
	}
}

func TestHashcash_GenerateChallenge(t *testing.T) {
	h := NewHashcash(8)
	challenge := h.GenerateChallenge()

	if len(challenge) != NonceSize {
		t.Errorf("challenge length = %d, want %d", len(challenge), NonceSize)
	}
}

func TestHashcash_Verify(t *testing.T) {
	t.Run("valid proof", func(t *testing.T) {
		h := NewHashcash(1)
		seed := []byte("test")
		proof := findValidProof(h, seed)

		if !h.Verify(seed, proof) {
			t.Errorf("Verify() = false, want true for valid proof")
		}
	})

	t.Run("invalid proof", func(t *testing.T) {
		h := NewHashcash(16)
		seed := []byte("test")

		invalidProof := make([]byte, NonceSize)
		invalidProof[0] = 0xFF

		if h.Verify(seed, invalidProof) {
			t.Errorf("Verify() = true, want false for invalid proof")
		}
	})
}

func TestHashcash_Solve(t *testing.T) {
	h := NewHashcash(1)
	challenge := h.GenerateChallenge()

	ctx := context.Background()
	proof, err := h.Solve(ctx, challenge)

	if err != nil {
		t.Fatalf("Solve returned an error: %v", err)
	}

	if !h.Verify(challenge, proof) {
		t.Errorf("Solve() returned invalid proof")
	}

	if len(proof) != NonceSize {
		t.Errorf("proof length = %d, want %d", len(proof), NonceSize)
	}
}

// Helper function to find a valid proof
func findValidProof(h *Hashcash, seed []byte) []byte {
	proof := make([]byte, NonceSize)
	counter := uint64(0)

	for {
		binary.LittleEndian.PutUint64(proof, counter)
		if h.Verify(seed, proof) {
			return proof
		}
		counter++
	}
}
