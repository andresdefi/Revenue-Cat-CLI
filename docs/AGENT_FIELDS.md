# Agent Field Presets

Use `--fields default` on these read commands to return the default preset for agent and script use. Use `--fields <csv>` to pick any available JSON fields from the matching API type.

`--agent` and `RC_AGENT=1` apply compact JSON, `--fields default`, and no next-step hints unless an explicit flag overrides that choice.

Fields listed under "All available fields" come from the current `internal/api/types.go` JSON tags. Invalid fields from the initial preset proposal were dropped: `customers entitlements` uses `expires_at` instead of `expires_date` and does not include `product_id`; webhooks do not include `status`; paywalls only expose `id` and `created_at`; virtual currencies do not include `id`; audit logs do not include `resource_type`.

## Projects

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc projects list` | `id,name,created_at` | `object,id,name,created_at,icon_url,icon_url_large` |

## Apps

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc apps list` | `id,name,type,project_id` | `object,id,name,type,created_at,project_id,app_store,mac_app_store,play_store` |
| `rc apps get` | `id,name,type,project_id` | `object,id,name,type,created_at,project_id,app_store,mac_app_store,play_store` |

## Products

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc products list` | `id,store_identifier,type,state,display_name,app_id` | `object,id,store_identifier,type,state,display_name,created_at,app_id,app,subscription,one_time` |
| `rc products get` | `id,store_identifier,type,state,display_name,app_id` | `object,id,store_identifier,type,state,display_name,created_at,app_id,app,subscription,one_time` |

## Entitlements

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc entitlements list` | `id,lookup_key,display_name,state` | `object,id,project_id,lookup_key,display_name,state,created_at,products` |
| `rc entitlements get` | `id,lookup_key,display_name,state` | `object,id,project_id,lookup_key,display_name,state,created_at,products` |

## Offerings

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc offerings list` | `id,lookup_key,display_name,is_current,state` | `object,id,project_id,lookup_key,display_name,is_current,state,created_at,metadata,packages` |
| `rc offerings get` | `id,lookup_key,display_name,is_current,state` | `object,id,project_id,lookup_key,display_name,is_current,state,created_at,metadata,packages` |

## Packages

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc packages list` | `id,lookup_key,display_name,position` | `object,id,lookup_key,display_name,position,created_at,products` |
| `rc packages get` | `id,lookup_key,display_name,position` | `object,id,lookup_key,display_name,position,created_at,products` |

## Customers

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc customers list` | `id,project_id,last_seen_at,last_seen_platform` | `object,id,project_id,first_seen_at,last_seen_at,last_seen_app_version,last_seen_country,last_seen_platform,last_seen_platform_version,active_entitlements,experiment` |
| `rc customers lookup` | `id,project_id,last_seen_at,last_seen_platform` | `object,id,project_id,first_seen_at,last_seen_at,last_seen_app_version,last_seen_country,last_seen_platform,last_seen_platform_version,active_entitlements,experiment` |
| `rc customers entitlements` | `entitlement_id,expires_at` | `object,entitlement_id,expires_at` |

## Subscriptions

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc subscriptions list` | `id,customer_id,status,current_period_ends_at,product_id,store` | `object,id,customer_id,original_customer_id,product_id,starts_at,current_period_starts_at,current_period_ends_at,ends_at,gives_access,pending_payment,auto_renewal_status,status,total_revenue_in_usd,presented_offering_id,environment,store,store_subscription_identifier,ownership,country,management_url` |
| `rc subscriptions get` | `id,customer_id,status,current_period_ends_at,product_id,store` | `object,id,customer_id,original_customer_id,product_id,starts_at,current_period_starts_at,current_period_ends_at,ends_at,gives_access,pending_payment,auto_renewal_status,status,total_revenue_in_usd,presented_offering_id,environment,store,store_subscription_identifier,ownership,country,management_url` |

## Purchases

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc purchases list` | `id,customer_id,purchased_at,product_id,store` | `object,id,customer_id,original_customer_id,product_id,purchased_at,revenue_in_usd,quantity,status,presented_offering_id,environment,store,store_purchase_identifier,ownership,country` |
| `rc purchases get` | `id,customer_id,purchased_at,product_id,store` | `object,id,customer_id,original_customer_id,product_id,purchased_at,revenue_in_usd,quantity,status,presented_offering_id,environment,store,store_purchase_identifier,ownership,country` |

## Webhooks

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc webhooks list` | `id,url` | `object,id,name,url,created_at` |
| `rc webhooks get` | `id,url` | `object,id,name,url,created_at` |

## Paywalls

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc paywalls list` | `id` | `object,id,created_at` |
| `rc paywalls get` | `id` | `object,id,created_at` |

## Virtual Currencies

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc currencies list` | `code,name,state` | `object,code,name,state,created_at` |
| `rc currencies get` | `code,name,state` | `object,code,name,state,created_at` |

## Audit Logs

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc audit-logs list` | `id,created_at,actor,action` | `object,id,action,actor,created_at,details` |

## Collaborators

| Command | Default preset | All available fields |
|---------|----------------|----------------------|
| `rc collaborators list` | `id,email,role` | `object,id,email,role` |
