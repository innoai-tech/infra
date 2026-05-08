# Go 工具链入口
[group: 'toolchain']
mod go 'tool/go/justfile'

# 示例应用入口
[group: 'app']
mod example 'internal/example/justfile'

# 列出所有可用命令
[group('meta')]
default:
    @just --list --list-submodules
