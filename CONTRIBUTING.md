# Contributing to BuildSift

Thank you for helping turn noisy failures into reliable, reusable diagnoses.

## Start with evidence

The best rule contribution begins with a real failure and a confirmed root cause. Before opening an issue or pull request:

1. Reduce the log to the smallest sample that still reproduces the diagnosis.
2. Remove tokens, credentials, private hostnames, usernames, customer data, and proprietary source code.
3. State the expected root cause in one sentence.
4. Explain which lines prove that conclusion and which later errors are only consequences.

Do not submit a log you do not have permission to publish.

## Development setup

BuildSift requires Go. From the repository root:

```bash
go test ./...
go vet ./...
go run ./cmd/buildsift --help
```

Format changed Go files before submitting:

```bash
gofmt -w path/to/changed.go
```

## Adding or changing a diagnosis

- Add a sanitized log fixture and its expected diagnosis first.
- Keep matching rules narrow enough to avoid unrelated logs.
- Prefer specific evidence such as a missing path or named exception over generic words such as `error` or `failed`.
- Test both the intended match and at least one nearby non-match.
- Preserve line references and deterministic output.
- Return an unknown result when the evidence is insufficient.

Avoid adding speculative remediation. A next-check command should gather evidence or reproduce the failing step; it should not make destructive changes.

## Pull requests

Keep each pull request focused on one diagnosis, parser, or infrastructure change. Include:

- The failure type and affected tool.
- A short explanation of the confirmed root cause.
- The sanitized fixture or a synthetic equivalent.
- Tests showing the behavior before and after the change.
- Any known false-positive boundary.

Run `go test ./...` and `go vet ./...` before opening the pull request. CI runs the same checks on Linux, macOS, and Windows.

## Bug reports

A useful bug report includes the BuildSift version, operating system, input mode, actual output, expected output, and the smallest sanitized log that demonstrates the problem.

If the report may expose a vulnerability or secret, follow [SECURITY.md](SECURITY.md) instead of opening a public issue.
