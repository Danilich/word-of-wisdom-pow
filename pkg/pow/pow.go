package pow

import "context"

type Pow interface {
	GenerateChallenge() []byte
	Verify(seed, proof []byte) bool
	Solve(ctx context.Context, challenge []byte) ([]byte, error)
}
