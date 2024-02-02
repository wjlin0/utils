package interactshutils

import (
	"github.com/pkg/errors"
	"time"
)

const (
	defaultInteractionDuration  = time.Minute
	defaultMaxInteractionsCount = 5000
	DefaultTimeout              = 10 * time.Second
)

var (
	ErrInteractshClientNotInitialized = errors.New("interactsh client not initialized")
)
