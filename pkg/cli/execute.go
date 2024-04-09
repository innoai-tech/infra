package cli

import (
	"context"
	"fmt"
	"os"
)

func Exec(ctx context.Context, app Command) {
	if err := Execute(ctx, app, os.Args[1:]); err != nil {
		fmt.Printf("%s\n", err)
		fmt.Printf("%#v\n", err)
		os.Exit(1)
	}
}

func Execute(ctx context.Context, c Command, args []string) error {
	if inputs, ok := c.(interface{ ParseArgs(args []string) }); ok {
		inputs.ParseArgs(args)
	}

	if e, ok := c.(interface {
		ExecuteContext(ctx context.Context) error
	}); ok {
		return e.ExecuteContext(ctx)
	}
	return nil
}
