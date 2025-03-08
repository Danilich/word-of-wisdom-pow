package pow

type Pow interface {
	GenerateChallenge() []byte
	Verify(seed, proof []byte) bool
	Solve(challenge []byte) []byte
}
