# Go 工具链入口
mod go 'tool/go/justfile'

# 示例代码
mod example 'internal/example/justfile'

# 列出所有可用命令
[group('meta')]
default:
    @just --list --list-submodules
