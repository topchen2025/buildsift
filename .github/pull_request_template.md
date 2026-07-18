## What failed?

Describe the build tool, confirmed root cause, and misleading downstream noise.

## What changed?

Explain the narrow rule, parser, fixture, documentation, or infrastructure change.

## Evidence

- [ ] Added or updated a sanitized fixture.
- [ ] Added a positive test and a nearby non-match where relevant.
- [ ] Confirmed the expected root cause against the original failure.
- [ ] Ran `go test ./...` and `go vet ./...`.
- [ ] Verified that no secrets, private hosts, usernames, customer data, or proprietary source are included.
