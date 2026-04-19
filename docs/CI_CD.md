# CI/CD

`rc` is designed to work in non-interactive shells with explicit output formats.

## Authentication

Use a RevenueCat API v2 secret key in your CI secret store.

```bash
export RC_BYPASS_KEYCHAIN=1
printf '%s\n' "$RC_API_KEY" | rc auth login --profile ci
rc projects set-default "$RC_PROJECT_ID" --profile ci
rc doctor --profile ci
```

For commands in scripts, prefer explicit profile and output flags:

```bash
rc products list --profile ci --project "$RC_PROJECT_ID" --output json
```

For compact, agent-friendly JSON in non-interactive jobs, use `--agent` or `RC_AGENT=1`. Agent mode applies compact JSON, the command's default field preset, and suppresses next-step hints unless an explicit flag overrides the output or fields.

```bash
rc --agent products list --profile ci --project "$RC_PROJECT_ID"
RC_AGENT=1 rc products list --profile ci --project "$RC_PROJECT_ID"
```

## GitHub Actions Example

```yaml
name: RevenueCat check

on:
  workflow_dispatch:

jobs:
  revenuecat:
    runs-on: ubuntu-latest
    steps:
      - name: Install rc
        run: curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh

      - name: Validate RevenueCat access
        env:
          RC_API_KEY: ${{ secrets.REVENUECAT_API_KEY }}
          RC_PROJECT_ID: ${{ vars.REVENUECAT_PROJECT_ID }}
          RC_BYPASS_KEYCHAIN: "1"
        run: |
          printf '%s\n' "$RC_API_KEY" | rc auth login --profile ci
          rc projects set-default "$RC_PROJECT_ID" --profile ci
          rc doctor --profile ci
          rc products list --profile ci --output json
```

## Output

- Use `--output json` for machine-readable scripts.
- Use `--pretty` when storing JSON logs for humans.
- Use `--quiet` to suppress progress and success messages.
- Use `--fields` to trim JSON output to the fields a job needs.
- Use `--fields default` to apply the command's preset from [AGENT_FIELDS.md](AGENT_FIELDS.md).
