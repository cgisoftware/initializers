package pacific

import "context"

type PacificHttpRepository interface {
	Send(context.Context, string, PacificInput) ([]byte, *PacificError)
}
