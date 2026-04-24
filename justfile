
mod go 'tool/go/justfile'
mod example 'internal/example/justfile'

# List stable repository entrypoints.
default:
    just --list --list-submodules
