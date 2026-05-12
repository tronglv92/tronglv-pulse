package consumer

import "context"

type (
	MessageHandler interface {
		Consume(ctx context.Context, payload MessageContext, args map[string]any) error
	}
)
