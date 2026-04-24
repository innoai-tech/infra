package cli

import (
	"context"
	"fmt"
	"os"
)

// Exec 使用进程参数执行命令，失败时直接退出进程。
func Exec(ctx context.Context, app Command) {
	if err := Execute(ctx, app, os.Args[1:]); err != nil {
		fmt.Printf("%s\n", err)
		fmt.Printf("%#v\n", err)
		os.Exit(1)
	}
}

// Execute 使用给定参数执行命令。
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
