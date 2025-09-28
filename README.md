# Infra

## Code Structure

```
cmd/
    <app_name>/
        main.go
        <sub_cmd>.go    
internal/                   # internal pkgs
pkg/                        # shared pkgs       
```

## Singleton in Context

```mermaid
flowchart TB
    classDef default fill:#fff;
    classDef sequential fill:#ffd2b3,stroke:#000;
    classDef parallel fill:#ffb3b3,stroke:#000;
    
    cli_run["cli exec"]
    shutdown_signals["shutdown\nsignals"]
    
    singletons[["singletons"]]:::sequential
    
    context(("ctx")):::sequential
    singleton_runner("Run"):::sequential
    singleton_server("Server"):::parallel
    
    cli_run
    ==> singletons
    ==> |"each singleton"| singletons_empty(( ))
    ==> |"unmarshal from cli flags"| singletons_with_flag_values(( ))
    ==> |"unmarshal from env vars"| singletons_with_env_vars(( ))
    ==> |"Init"| singletons
    
    singletons
    --> |"after init"| singletons_init(( ))
    --> |"inject singletons\n with ContextInjector"| context
    
    context
    --> |"each singleton"| singleton_runner_check(( )) 
    
    singleton_runner_check
    --> |"if can Run"| singleton_runner
    --> singleton_after_run(( ))
    
    singleton_runner_check
    --> singleton_after_run(( ))
    
    singleton_after_run
    --> singleton_server_check(( ))
    
    singleton_server_check
    --> |"if can Serve"| singleton_server
    --> |"parallel serve"| singleton_server_running(( ))
    
    singleton_server_check
    --> |"exit"| exit(( ))
    
    shutdown_signals
    -..->  singleton_server_running
    -..->  |"parallel shutdown\ngracefully"| exit(( ))  
      
    
    singleton_server_running(("Running Server")):::parallel
    context_server(("ctx")):::parallel
    server_task("handler"):::parallel
    
    singleton_server_running
    -..->|"on trigger\n request / event"| start(( ))
    -->  |"inject singletons\n with ContextInjector"| context_server
    -->  |"exec"| server_task
    -..->|"done"| done(( ))
    
```

```go
package main

import (
	"context"

	"github.com/octohelm/x/logr"
	"github.com/innoai-tech/infra/pkg/otel"
	"github.com/innoai-tech/infra/pkg/cli"
)

var App = cli.NewApp("app", "0.0.0")

func init() {
	cli.AddTo(App, &Server{})
}

type Server struct {
	// declare as sub command
	// tag component is special set
	// when define, --dump-k8s added, for dump kube app component cue files 
	cli.C `component:"server"`

	// singleton
	// with interface { InjectContext() context.Context },
	// we could get singleton from context
	otel.Otel

	// Runner singleton
	// run once when cli exec
	Runner

	// Server singleton
	// start serve when cli exec
	// if multi servers exists, will parallel serve.
	ServerOrAgent
}

type Runner struct {
}

func (r *Runner) Run(ctx context.Context) error {
	// get singleton from ctx
	logr.FromContext(ctx).Info("run")
	return nil
}

type ServerOrAgent struct {
	// expose could be http/tcp/udp
	// generating services into cue files only when this setting defined.
	Addr string `flag:",omitempty,expose=http"`
}

func (r *ServerOrAgent) Serve(ctx context.Context) error {
	return nil
}

func (r *ServerOrAgent) Shutdown(ctx context.Context) error {
	return nil
}
```