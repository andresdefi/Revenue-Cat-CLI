# Support

## Getting Help

- **Bug reports**: [Open an issue](https://github.com/andresdefi/Revenue-Cat-CLI/issues/new?template=bug_report.yml)
- **Feature requests**: [Open an issue](https://github.com/andresdefi/Revenue-Cat-CLI/issues/new?template=feature_request.yml)
- **Questions & discussion**: [GitHub Discussions](https://github.com/andresdefi/Revenue-Cat-CLI/discussions)

## Before Opening an Issue

1. Check existing [issues](https://github.com/andresdefi/Revenue-Cat-CLI/issues) and [discussions](https://github.com/andresdefi/Revenue-Cat-CLI/discussions)
2. Run `rc version` and include the output
3. Try with `--output json` to see the raw API response
4. Include the full error message

## Troubleshooting

### Authentication

```bash
# Check if you're logged in
rc auth status

# Re-login if needed
rc auth logout
rc auth login
```

### "no project specified"

```bash
# Set a default project
rc projects list
rc projects set-default <project-id>
```

### API errors

RevenueCat API errors include a `doc_url` field. Check the linked documentation for details on the specific error.

### Rate limits

If you see `rate_limit_error`, wait and retry. Rate limits:
- Project configuration: 60 requests/minute
- Customer information: 480 requests/minute
- Charts & metrics: 5 requests/minute

## RevenueCat Support

For issues with the RevenueCat service itself (not this CLI), contact [RevenueCat Support](https://www.revenuecat.com/support).
