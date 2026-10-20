# AI Usage

## Tools

- **Claude Code** (VS Code extension, Claude Opus 5.5) as a programming assistant.

## How it was used

The author organized the project: wrote the plan in `CLAUDE.md` (scope, stages, technical rules, code conventions, demo scenario, test requirements) and settled most design decisions beforehand (`Person` embedded by value, ids passed to constructors, `Attender` in episode records, `HireAttender`, `IsAvailable`, the `Hospital` orchestrating sleep attacks, the injectable clock). The author also chose the state values, the hospital name and the demo data, reviewed every stage before accepting it, and managed the Git history.

Claude worked within that plan, one stage at a time: it generated the code for each stage, wrote the tests, helped find and fix errors, and explained Go concepts the author did not know yet (package-level encapsulation, value vs. pointer embedding, sentinel errors with `%w`, compile-time interface checks).

## Decisive prompts

1. **"Start Stage 0 following CLAUDE.md"**: every request was answered against the author's plan, which kept the code consistent across stages.
2. **"Don't ask me technical questions, use the best practices and the instructions in the claude.md"**: Claude made the implementation details and had to justify each one in writing, which the author used to study for the defense.
3. **"If there is something creative that I can configure tell me"**: design was left to the rules, and presentation choices to the author.

## Incorrect or non-idiomatic code and corrections

1. **Exported `Patient.WakeUp`.** It let a patient wake up without freeing their room. Fixed by making it unexported; waking a patient now goes only through `Hospital.WakePatient`, which also releases the bed.
2. **Dead code for an unused import.** `var _ = errors.Is` was added only to make an unused import compile. Fixed by removing both.
3. **Unused test variable and misleading test name.** `_ = awake` silenced an unused variable, and a test's name said the opposite of what it checked. Fixed by removing the variable and renaming the test.

## Learnings

- In Go, encapsulation is per package: an unexported method protects an invariant; a comment does not.
- Polymorphism only means something when the implementations behave differently.
- Injecting the clock makes date-dependent code testable.
- Generated code must be read, not just run: all three errors compiled and passed the checks.
