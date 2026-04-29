# DevSecOps Instructions

This repo uses Trivy and Semgrep for local and CI security checks.

## Tools

Install locally:

```bash
brew install trivy semgrep
```

Optional Homebrew noise reduction:

```bash
export HOMEBREW_NO_INSTALL_CLEANUP=1
export HOMEBREW_NO_ENV_HINTS=1
```

Add those exports to `~/.zshrc` only if you want them permanently.

## Secrets

Do not commit Semgrep tokens, Supabase passwords, API keys, `.env`, or `.env.local`.

If a Semgrep token is pasted into logs, chat, or a tracked file, rotate it in Semgrep Cloud and run:

```bash
semgrep login
```

Semgrep stores local auth in `~/.semgrep/settings.yml`, outside this repo.

## Local Scans

Run all local security scans:

```bash
make security
```

Equivalent direct commands:

```bash
trivy fs --config trivy.yaml .
semgrep ci
```

`make security` skips missing tools and skips Semgrep if the CLI is not logged in.

The script also sets:

- `DOCKER_CONFIG=/tmp/life3-empty-docker-config` by default, so Trivy does not trip over local Docker credential helpers when downloading its public vulnerability database.
- `TRIVY_DB_REPOSITORY=ghcr.io/aquasecurity/trivy-db` and `TRIVY_JAVA_DB_REPOSITORY=ghcr.io/aquasecurity/trivy-java-db` by default.
- `SSL_CERT_FILE` from Python `certifi` when available, which helps Semgrep on Homebrew Python installs with empty system trust anchors.

## Trivy

Tracked files:

- `.github/workflows/trivy-ci-scan.yml`
- `trivy.yaml`
- `config/trivy/secret.yaml`
- `.trivyignore`

The Trivy scan covers:

- dependency vulnerabilities
- secret detection
- IaC/misconfiguration checks

The CI workflow writes SARIF and uploads it to GitHub code scanning. It gates the workflow on `HIGH` and `CRITICAL` findings while ignoring unfixed vulnerabilities.

Use `.trivyignore` only for accepted findings. Each ignored ID should have a short justification in the PR or commit message.

## Semgrep

Local authenticated scan:

```bash
semgrep ci
```

Semgrep Cloud login enables additional proprietary registry rules and custom org policies. Keep policy changes in Semgrep Cloud, not in this repo, unless a project-local Semgrep rule is intentionally added later.

If Semgrep fails locally with a log-file permission error inside a sandboxed agent, run it from your normal terminal session. The token and log file live under `~/.semgrep`, outside the repository.

## CI Expectations

Before merging security-sensitive changes, run:

```bash
make test
make lint
make security
```

If a finding is real, fix it. If it is a false positive, prefer a narrow ignore rule near the scanner config rather than broad path exclusions.
