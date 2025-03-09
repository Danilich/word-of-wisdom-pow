package services

import (
	"word-of-wisdom-pow/pkg/pow"
)

// PowService is a proof of work service
type PowService struct {
	pow pow.Pow
}

func NewPowService(pow pow.Pow) *PowService {
	return &PowService{
		pow: pow,
	}
}

// GenerateChallenge generates a new pow challenge
func (s *PowService) GenerateChallenge() []byte {
	return s.pow.GenerateChallenge()
}

// VerifyProof verifies a pow solution
func (s *PowService) VerifyProof(seed, proof []byte) bool {
	return s.pow.Verify(seed, proof)
}

// CreateDefaultPow creates a default challenge
func CreateDefaultPow() *PowService {
	return NewPowService(pow.NewDefaultHashcash())
}

// CreatePow creates a challenge with difficulty
func CreatePow(difficulty uint8) *PowService {
	return NewPowService(pow.NewHashcash(difficulty))
}
