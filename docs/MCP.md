# RevenueCat MCP Server

`rc mcp serve` starts an experimental Model Context Protocol server over stdio.
It exposes a focused subset of RevenueCat operations as structured tools for
agents. Use it when an agent needs JSON-shaped RevenueCat context or a small
set of safe project/customer actions without shelling out to the full CLI.

## Authentication

The server resolves credentials in this order:

1. `RC_API_KEY`
2. Stored CLI authentication from `rc auth login`

The API key still needs the RevenueCat permissions required by each underlying
API endpoint.

## Tools

| Tool | Kind | RevenueCat API surface |
|------|------|------------------------|
| `list_projects` | read | `GET /projects` |
| `list_products` | read | `GET /projects/{project_id}/products` |
| `get_product` | read | `GET /projects/{project_id}/products/{product_id}` |
| `create_product` | write | `POST /projects/{project_id}/products` |
| `list_entitlements` | read | `GET /projects/{project_id}/entitlements` |
| `get_entitlement` | read | `GET /projects/{project_id}/entitlements/{entitlement_id}` |
| `create_entitlement` | write | `POST /projects/{project_id}/entitlements` |
| `list_offerings` | read | `GET /projects/{project_id}/offerings` |
| `get_offering` | read | `GET /projects/{project_id}/offerings/{offering_id}` |
| `lookup_customer` | read | `GET /projects/{project_id}/customers/{customer_id}` |
| `list_customer_entitlements` | read | `GET /projects/{project_id}/customers/{customer_id}/active_entitlements` |
| `grant_entitlement` | write | `POST /projects/{project_id}/customers/{customer_id}/actions/grant_entitlement` |
| `list_subscriptions` | read | `GET /projects/{project_id}/subscriptions` |
| `get_subscription` | read | `GET /projects/{project_id}/subscriptions/{subscription_id}` |
| `metrics_overview` | read | `GET /projects/{project_id}/metrics/overview` |

## CLI vs MCP

Prefer the CLI for interactive workflows, shell scripts, generated docs, and
commands that are not exposed as MCP tools. Prefer MCP when an agent needs a
small set of structured RevenueCat calls in one session.

The MCP server intentionally exposes fewer tools than the CLI. The full command
surface remains available through `rc --help` and `docs/COMMANDS.md`.

## Example

```bash
RC_API_KEY=sk_test_xxx rc mcp serve
```

Claude Code configuration:

```json
{
  "mcpServers": {
    "revenuecat": {
      "command": "rc",
      "args": ["mcp", "serve"]
    }
  }
}
```
