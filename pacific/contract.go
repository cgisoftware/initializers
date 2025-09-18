package pacific

import (
	"context"
	"time"
)

type PacificHttpRepository interface {
	Send(context.Context, string, PacificInput, time.Duration) ([]byte, error)
}
