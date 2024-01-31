package interactshutils

import (
	"github.com/pkg/errors"
	"time"
)

const (
	defaultInteractionDuration  = time.Minute
	defaultMaxInteractionsCount = 5000
)

var (
	ErrInteractshClientNotInitialized = errors.New("interactsh client not initialized")
)
