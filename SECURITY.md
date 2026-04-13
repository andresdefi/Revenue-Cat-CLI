# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in `rc`, please report it responsibly using [GitHub Security Advisories](https://github.com/andresdefi/Revenue-Cat-CLI/security/advisories/new).

**Do not open a public issue for security vulnerabilities.**

We will acknowledge your report within 48 hours and work with you to understand and address the issue.

## Scope

- The `rc` CLI binary and its dependencies
- Authentication token storage (keychain and config file)
- API request handling and data transmission

## Out of Scope

- The RevenueCat API itself (report to [RevenueCat](https://www.revenuecat.com/security))
- Third-party dependencies (report to the upstream project)

## Best Practices

- Always use `sk_` secret keys with minimal required permissions
- Never commit API keys to version control
- Use the system keychain for token storage when available
- Keep `rc` updated to the latest version
