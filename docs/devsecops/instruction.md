# DevSecOps Instructions

This repo uses Trivy, Semgrep, and optional Snyk checks for local and CI security coverage.

## Tools

Install locally:

```bash
brew install trivy semgrep snyk
```

Optional Homebrew noise reduction:

```bash
export HOMEBREW_NO_INSTALL_CLEANUP=1
export HOMEBREW_NO_ENV_HINTS=1
```

Add those exports to `~/.zshrc` only if you want them permanently.

## Secrets

Do not commit Semgrep tokens, Snyk tokens, Supabase passwords, API keys, `.env`, or `.env.local`.

If a Semgrep token is pasted into logs, chat, or a tracked file, rotate it in Semgrep Cloud and run:

```bash
semgrep login
```

Semgrep stores local auth in `~/.semgrep/settings.yml`, outside this repo. Snyk stores local auth outside the repo as well.

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
- `SEMGREP_LOG_FILE=/tmp/life3-semgrep.log` by default, so sandboxed agents do not need write access to `~/.semgrep/semgrep.log`.

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

Tracked files:

- `.github/workflows/semgrep-ci-scan.yml`

Local authenticated scan:

```bash
semgrep ci
```

Semgrep Cloud login enables additional proprietary registry rules and custom org policies. Keep policy changes in Semgrep Cloud, not in this repo, unless a project-local Semgrep rule is intentionally added later.

If Semgrep cannot reach `semgrep.dev`, rerun from a network-enabled terminal or CI runner. The token lives under `~/.semgrep`, outside the repository.

In GitHub Actions, Semgrep uses `SEMGREP_APP_TOKEN` when that repository secret is configured. Without the secret, the workflow falls back to Semgrep Community Edition with `semgrep scan --config auto --metrics=auto`. Semgrep requires metrics in `auto` mode because it uses telemetry to select and improve registry rules; use a checked-in local Semgrep rules file instead if CI must run with metrics disabled.

## Snyk

Tracked files:

- `.github/workflows/snyk-ci-scan.yml`

Snyk requires a repository secret named `SNYK_TOKEN`. The GitHub Actions workflow skips Snyk when the token is not configured, which keeps pull requests from forks from failing only because secrets are unavailable.

Local authenticated scan:

```bash
snyk auth
snyk test --all-projects --detection-depth=6 --severity-threshold=high
```

## CI Expectations

Before merging security-sensitive changes, run:

```bash
make test
make lint
make security
```

If a finding is real, fix it. If it is a false positive, prefer a narrow ignore rule near the scanner config rather than broad path exclusions.
