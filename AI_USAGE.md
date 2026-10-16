# AI Usage

> Work in progress. This log is updated at every stage and finalized in stage 9.

## Tools

- Claude Code (VS Code extension), model Claude Opus 5.5.

## Decisive prompts

- Stage 0: "Start Stage 0 following CLAUDE.md". The project plan, rules and conventions were written beforehand in `CLAUDE.md` so that every stage is generated against the same constraints.

## Incorrect or non-idiomatic generated code and corrections

<!-- Format: stage, what was generated, why it was wrong, correction. -->

## Other issues found while using the tool

- Stage 0: running `go mod edit -go=1.21` from PowerShell failed because PowerShell split the `-go=` argument. `go.mod` was written directly instead with `go 1.21`, so the project builds on any Go 1.21+ toolchain rather than requiring the locally installed 1.27.

## Learnings
