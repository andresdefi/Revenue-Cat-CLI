# RevenueCat API Coverage

Generated: 2026-05-06T11:24:46Z

Spec source: `https://www.revenuecat.com/docs/redocusaurus/plugin-redoc-0.yaml`

Summary: 96/96 implemented, 81/96 tested, 96/96 documented.

| Method | Endpoint | Implemented | Tested | Documented |
|--------|----------|-------------|--------|------------|
| DELETE | `/projects/{project_id}/apps/{app_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/customers/{customer_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/entitlements/{entitlement_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/integrations/webhooks/{webhook_integration_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/offerings/{offering_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/packages/{package_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/paywalls/{paywall_id}` | yes | yes | yes |
| DELETE | `/projects/{project_id}/products/{product_id}` | yes | no | yes |
| DELETE | `/projects/{project_id}/virtual_currencies/{virtual_currency_code}` | yes | yes | yes |
| GET | `/projects` | yes | yes | yes |
| GET | `/projects/{project_id}/apps` | yes | yes | yes |
| GET | `/projects/{project_id}/apps/{app_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/apps/{app_id}/public_api_keys` | yes | no | yes |
| GET | `/projects/{project_id}/apps/{app_id}/store_kit_config` | yes | no | yes |
| GET | `/projects/{project_id}/audit_logs` | yes | yes | yes |
| GET | `/projects/{project_id}/charts/{chart_name}` | yes | yes | yes |
| GET | `/projects/{project_id}/charts/{chart_name}/options` | yes | yes | yes |
| GET | `/projects/{project_id}/collaborators` | yes | yes | yes |
| GET | `/projects/{project_id}/customers` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/active_entitlements` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/aliases` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/attributes` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/invoices` | yes | no | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/invoices/{invoice_id}/file` | yes | no | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/purchases` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/subscriptions` | yes | yes | yes |
| GET | `/projects/{project_id}/customers/{customer_id}/virtual_currencies` | yes | yes | yes |
| GET | `/projects/{project_id}/entitlements` | yes | yes | yes |
| GET | `/projects/{project_id}/entitlements/{entitlement_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/entitlements/{entitlement_id}/products` | yes | yes | yes |
| GET | `/projects/{project_id}/integrations/webhooks` | yes | yes | yes |
| GET | `/projects/{project_id}/integrations/webhooks/{webhook_integration_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/metrics/overview` | yes | yes | yes |
| GET | `/projects/{project_id}/offerings` | yes | yes | yes |
| GET | `/projects/{project_id}/offerings/{offering_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/offerings/{offering_id}/packages` | yes | yes | yes |
| GET | `/projects/{project_id}/packages/{package_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/packages/{package_id}/products` | yes | yes | yes |
| GET | `/projects/{project_id}/paywalls` | yes | yes | yes |
| GET | `/projects/{project_id}/paywalls/{paywall_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/products` | yes | yes | yes |
| GET | `/projects/{project_id}/products/{product_id}` | yes | no | yes |
| GET | `/projects/{project_id}/purchases` | yes | yes | yes |
| GET | `/projects/{project_id}/purchases/{purchase_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/purchases/{purchase_id}/entitlements` | yes | yes | yes |
| GET | `/projects/{project_id}/subscriptions` | yes | yes | yes |
| GET | `/projects/{project_id}/subscriptions/{subscription_id}` | yes | yes | yes |
| GET | `/projects/{project_id}/subscriptions/{subscription_id}/authenticated_management_url` | yes | yes | yes |
| GET | `/projects/{project_id}/subscriptions/{subscription_id}/entitlements` | yes | yes | yes |
| GET | `/projects/{project_id}/subscriptions/{subscription_id}/transactions` | yes | yes | yes |
| GET | `/projects/{project_id}/virtual_currencies` | yes | yes | yes |
| GET | `/projects/{project_id}/virtual_currencies/{virtual_currency_code}` | yes | yes | yes |
| POST | `/projects` | yes | yes | yes |
| POST | `/projects/{project_id}/apps` | yes | yes | yes |
| POST | `/projects/{project_id}/apps/{app_id}` | yes | no | yes |
| POST | `/projects/{project_id}/customers` | yes | yes | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/actions/assign_offering` | yes | no | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/actions/grant_entitlement` | yes | yes | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/actions/restore_purchase_by_order_id` | yes | no | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/actions/revoke_granted_entitlement` | yes | yes | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/actions/transfer` | yes | no | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/attributes` | yes | yes | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/virtual_currencies/transactions` | yes | yes | yes |
| POST | `/projects/{project_id}/customers/{customer_id}/virtual_currencies/update_balance` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements/{entitlement_id}` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements/{entitlement_id}/actions/archive` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements/{entitlement_id}/actions/attach_products` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements/{entitlement_id}/actions/detach_products` | yes | yes | yes |
| POST | `/projects/{project_id}/entitlements/{entitlement_id}/actions/unarchive` | yes | no | yes |
| POST | `/projects/{project_id}/integrations/webhooks` | yes | yes | yes |
| POST | `/projects/{project_id}/integrations/webhooks/{webhook_integration_id}` | yes | yes | yes |
| POST | `/projects/{project_id}/offerings` | yes | yes | yes |
| POST | `/projects/{project_id}/offerings/{offering_id}` | yes | yes | yes |
| POST | `/projects/{project_id}/offerings/{offering_id}/actions/archive` | yes | yes | yes |
| POST | `/projects/{project_id}/offerings/{offering_id}/actions/unarchive` | yes | no | yes |
| POST | `/projects/{project_id}/offerings/{offering_id}/packages` | yes | yes | yes |
| POST | `/projects/{project_id}/packages/{package_id}` | yes | yes | yes |
| POST | `/projects/{project_id}/packages/{package_id}/actions/attach_products` | yes | yes | yes |
| POST | `/projects/{project_id}/packages/{package_id}/actions/detach_products` | yes | yes | yes |
| POST | `/projects/{project_id}/paywalls` | yes | yes | yes |
| POST | `/projects/{project_id}/products` | yes | yes | yes |
| POST | `/projects/{project_id}/products/{product_id}` | yes | no | yes |
| POST | `/projects/{project_id}/products/{product_id}/actions/archive` | yes | yes | yes |
| POST | `/projects/{project_id}/products/{product_id}/actions/unarchive` | yes | no | yes |
| POST | `/projects/{project_id}/products/{product_id}/create_in_store` | yes | no | yes |
| POST | `/projects/{project_id}/purchases/{purchase_id}/actions/refund` | yes | yes | yes |
| POST | `/projects/{project_id}/subscriptions/{subscription_id}/actions/cancel` | yes | yes | yes |
| POST | `/projects/{project_id}/subscriptions/{subscription_id}/actions/extend` | yes | yes | yes |
| POST | `/projects/{project_id}/subscriptions/{subscription_id}/actions/refund` | yes | yes | yes |
| POST | `/projects/{project_id}/subscriptions/{subscription_id}/transactions/{transaction_id}/actions/refund` | yes | yes | yes |
| POST | `/projects/{project_id}/virtual_currencies` | yes | yes | yes |
| POST | `/projects/{project_id}/virtual_currencies/{virtual_currency_code}` | yes | yes | yes |
| POST | `/projects/{project_id}/virtual_currencies/{virtual_currency_code}/actions/archive` | yes | yes | yes |
| POST | `/projects/{project_id}/virtual_currencies/{virtual_currency_code}/actions/unarchive` | yes | yes | yes |

## Field Drift Warnings

- AppStoreApp missing local JSON field(s): app_store
- ChartData missing local JSON field(s): category, description, display_type, documentation_link, end_date, filtering_allowed, last_computed_at, measures, resolution, segmenting_allowed, segments, segments_limit, start_date, summary, unsupported_params, user_selectors, yaxis, yaxis_currency
- ChartOptions missing local JSON field(s): filters, resolutions, segments, user_selectors
- Collaborator missing local JSON field(s): accepted_at, has_mfa, name
- Customer missing local JSON field(s): attributes
- CustomerAlias missing local JSON field(s): created_at
- CustomerAttribute missing local JSON field(s): updated_at
- Invoice missing local JSON field(s): invoice_url, issued_at, line_items, paid_at, total_amount
- Paywall missing local JSON field(s): components, name, offering, offering_id, published_at
- PlayStoreApp missing local JSON field(s): play_store
- Purchase missing local JSON field(s): entitlements
- Subscription missing local JSON field(s): entitlements, pending_changes
- VirtualCurrency missing local JSON field(s): description, product_grants, project_id
