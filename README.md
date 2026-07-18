# BuildSift

**Your build failed. BuildSift finds the one error that actually matters.**

No AI. No uploads. Just an evidence-backed root cause from the logs already on your machine.

[简体中文](README.zh-CN.md)

![BuildSift turns a noisy build log into one root cause, evidence, and a next check](docs/hero.svg)

## Why BuildSift?

Build tools are excellent at producing logs and surprisingly bad at telling you which line started the failure. One missing file can become thousands of lines of stack traces, skipped modules, and secondary errors.

BuildSift removes that noise. It ranks concrete failure signals, collapses downstream cascades, and points back to the original evidence so you can verify every conclusion.

### Before

```text
[INFO] Reactor Summary for payments-parent 1.4.0:
[INFO] payments-api ................................ FAILURE
[INFO] payments-service ............................ SKIPPED
[INFO] payments-web ................................ SKIPPED
[ERROR] Failed to execute goal org.apache.maven.plugins:maven-pmd-plugin:...
... 2,184 more lines ...
[ERROR] Re-run Maven using the -X switch to enable full debug logging.
```

### After

```text
BUILDSIFT DIAGNOSIS
===================
ROOT CAUSE [HIGH · MAVEN]
NoSuchFileException: ~/work/quality/target/pmd.xml

EVIDENCE
  L1842  NoSuchFileException: ~/work/quality/target/pmd.xml

CASCADE
  17 additional failure signals folded

NEXT CHECK
  mvn -e -X
```

The example is illustrative; BuildSift always reports the evidence found in your own log.

## 30-second quick start

Install with Go:

```bash
go install github.com/topchen2025/buildsift/cmd/buildsift@latest
```

Or download the binary for your platform from [Releases](https://github.com/topchen2025/buildsift/releases/latest) and place it on your `PATH`. Then wrap a command:

```bash
buildsift -- mvn test
```

BuildSift streams the command output normally. If the command fails, it prints a short diagnosis and exits with the command's original status.

From a source checkout, replace the install command with:

```bash
go install ./cmd/buildsift
```

## Other inputs

Analyze a saved log:

```bash
buildsift build.log
```

Or pipe CI output directly:

```bash
gh run view --log-failed | buildsift
```

BuildSift v0.1 focuses on Maven, Gradle, npm/pnpm, and Docker/Compose failures.

## Why deterministic rules?

Build logs are evidence, not a creative-writing prompt.

- **Explainable:** every diagnosis points to the source lines that produced it.
- **Repeatable:** the same log and rule set produce the same result.
- **Fast:** no model download, API round trip, or account is required.
- **Honest about uncertainty:** low-confidence input is reported as unknown instead of guessed.
- **Easy to test:** every rule can ship with a sanitized log fixture and an expected result.

## Privacy by design

BuildSift analyzes logs locally and does not send them to a server. Its analyzer needs no API key and makes no network request. Wrapped build commands can still use the network exactly as they normally would.

Diagnostic evidence masks common token, password, URL-credential, and home-directory patterns. When BuildSift wraps a command, its original streamed output remains untouched. Redaction is best-effort, not a guarantee, so inspect anything before sharing it.

Logs can still contain credentials, private paths, source snippets, or customer data. Review and sanitize a log before attaching it to an issue or sharing it with another person. See [SECURITY.md](SECURITY.md) for responsible reporting.

## Design principles

1. Find the earliest specific cause, not the loudest final error.
2. Prefer evidence over speculation.
3. Collapse consequences without hiding the original log.
4. Preserve the wrapped command's output and exit status.
5. Keep the default experience local, fast, and dependency-free.

## Roadmap

- Expand high-quality fixtures for Maven, Gradle, npm/pnpm, and Docker/Compose.
- Add JSON and Markdown output for CI annotations and issue reports.
- Publish an official GitHub Action.
- Expand secret and private-path redaction coverage with adversarial fixtures.
- Grow community-maintained rule packs without turning the core into a plugin framework.

The roadmap is intentionally small. BuildSift should become more accurate before it becomes more configurable.

## Contributing

The highest-value contribution is a sanitized real-world failure log with the expected root cause. It turns one frustrating incident into a regression test that helps everyone.

Read [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow and fixture requirements. Please use GitHub Security Advisories for vulnerabilities rather than opening a public issue.

## License

BuildSift is available under the [MIT License](LICENSE).
