# AI Usage

> Work in progress. This log is updated at every stage and finalized in stage 9.

## Tools

- Claude Code (VS Code extension), model Claude Opus 5.5.

## Decisive prompts

- Stage 0: "Start Stage 0 following CLAUDE.md". The project plan, rules and conventions were written beforehand in `CLAUDE.md` so that every stage is generated against the same constraints.
- Stage 2: "Don't ask me technical questions, for these use the best practices and the instructions in the claude.md". From this point the tool decided implementation details itself and documented each decision in `CLAUDE.md` with its justification, which the author reviewed stage by stage.

## Incorrect or non-idiomatic generated code and corrections

<!-- Format: stage, what was generated, why it was wrong, correction. -->

1. **Exported `Patient.WakeUp` broke the room invariant (generated in stage 2, fixed in stage 6).** The generated `Patient` exposed `WakeUp()` publicly and only documented that releasing the room was "the Hospital's responsibility". Any caller (for example `main`) could wake a patient directly and leave their room marked `Occupied` by an awake patient. A comment cannot enforce an invariant. **Correction:** the method became unexported (`wakeUp`) and the only public entry point is `Hospital.WakePatient`, which wakes the patient and releases the bed in the same operation, the same pattern already used for `moveToBed` and `Room.occupy`/`release`.
2. **Dead code to keep an unused import (stage 6).** The first version of `hospital.go` imported `errors` without using it and added `var _ = errors.Is` with a misleading comment just to make the file compile. This is not idiomatic: an unused import should simply be removed. **Correction:** both the import and the blank assignment were deleted.

## Other issues found while using the tool

- Stage 0: running `go mod edit -go=1.21` from PowerShell failed because PowerShell split the `-go=` argument. `go.mod` was written directly instead with `go 1.21`, so the project builds on any Go 1.21+ toolchain rather than requiring the locally installed 1.27.

## Learnings
