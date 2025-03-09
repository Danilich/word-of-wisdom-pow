package pow

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/rand"
	"time"
)

const (
	// ByteSize represents the number of bits in a byte
	ByteSize = 8
	// NonceSize defines the size of the nonce in bytes
	NonceSize = 8
	// MaxByteValue represents all bits set in a byte (0xFF)
	MaxByteValue = 0xFF
	// DefaultDifficulty sets the default number of leading zeros
	DefaultDifficulty = 20
	// RecreateLimit defines the number of attempts after which the random generator will be recreated
	RecreateLimit = 1000000
)

type Hashcash struct {
	// Difficulty specifies the number of leading zero
	Difficulty uint8
	random     *rand.Rand
}

func NewDefaultHashcash() *Hashcash {
	return NewHashcash(DefaultDifficulty)
}

// NewHashcash creates a new Hashcash  with the specified difficulty
func NewHashcash(difficulty uint8) *Hashcash {
	if difficulty == 0 {
		difficulty = DefaultDifficulty
	}

	return &Hashcash{
		Difficulty: difficulty,
		random:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateChallenge creates a random challenge for the client to solve
func (pow *Hashcash) GenerateChallenge() []byte {
	challenge := make([]byte, NonceSize)
	binary.BigEndian.PutUint64(challenge, uint64(pow.random.Int63()))
	return challenge
}

// Verify checks if the provided proof solves the challenge
func (pow *Hashcash) Verify(challenge, proof []byte) bool {
	hash := calculateHash(challenge, proof)
	return hasLeadingZeros(hash, pow.Difficulty)
}

// Solve finds a proof that satisfies the challenge with the given difficulty
func (pow *Hashcash) Solve(ctx context.Context, challenge []byte) ([]byte, error) {
	var solution []byte
	var attempts uint64

	// Check if context is canceled
	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("solving was canceled")
		default:
		}

		attempts++
		solution = make([]byte, NonceSize)
		binary.BigEndian.PutUint64(solution, uint64(pow.random.Int63()))

		if pow.Verify(challenge, solution) {
			break
		}

		// Periodically recreate the random generator to avoid potential biases
		if attempts%RecreateLimit == 0 {
			pow.random = rand.New(rand.NewSource(time.Now().UnixNano()))
		}
	}

	return solution, nil
}

// calculateHash calculate hash of the challenge and proof
func calculateHash(challenge, proof []byte) []byte {
	hasher := sha256.New()
	hasher.Write(challenge)
	hasher.Write(proof)
	return hasher.Sum(nil)
}

// hasLeadingZeros checks if the hash has the required leading zeros
func hasLeadingZeros(hash []byte, difficulty uint8) bool {
	leadingZeroBytes := int(difficulty / ByteSize)
	leadingZeroBits := int(difficulty % ByteSize)

	for _, b := range hash[:leadingZeroBytes] {
		if b != 0 {
			return false
		}
	}

	if leadingZeroBits > 0 {
		mask := byte(MaxByteValue << (ByteSize - leadingZeroBits))
		return (hash[leadingZeroBytes] & mask) == 0
	}

	return true
}
