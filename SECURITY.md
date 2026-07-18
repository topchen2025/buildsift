# Security Policy

## Supported versions

Security fixes are applied to the latest released version.

| Version | Supported |
| --- | --- |
| Latest release | Yes |
| Older releases | No |

## Reporting a vulnerability

Please report vulnerabilities privately through GitHub Security Advisories for this repository. Do not include credentials, private logs, exploit details, or customer data in a public issue.

Include the affected version, platform, reproduction steps, impact, and a minimal proof of concept when possible. You should receive an acknowledgement within seven days. Please allow time for a fix and release before public disclosure.

Relevant security issues include command injection, unsafe handling of terminal control sequences, path traversal, unintended file access, exposure of log contents, and compromised release artifacts.

## Log safety

Build logs may contain tokens, environment variables, private URLs, source snippets, file paths, or customer data. BuildSift masks several common credential and home-directory patterns in diagnostic evidence, but it leaves a wrapped command's original streamed output untouched. Redaction is best-effort. Local processing does not make a log safe to publish. Review and sanitize logs before sharing them in issues, pull requests, chat, or generated reports.

BuildSift intentionally runs a command supplied after `--`. Only wrap commands you already trust, and review shell quoting exactly as you would without BuildSift.
