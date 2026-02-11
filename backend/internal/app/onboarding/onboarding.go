package onboarding

import (
	"github.com/sonalys/animeman/internal/ports"
)

type (
	onboarder struct {
		userRepository     ports.UserRepository
		prowlarrRepository ports.ProwlarrRepository
	}
)
