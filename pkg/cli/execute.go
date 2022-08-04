package cli

import "context"

func Execute(ctx context.Context, c Command, args []string) error {
	if inputs, ok := c.(interface {
		ParseArgs(args []string)
	}); ok {
		inputs.ParseArgs(args)
	}
	if e, ok := c.(interface {
		ExecuteContext(ctx context.Context) error
	}); ok {
		return e.ExecuteContext(ctx)
	}
	return nil
}
