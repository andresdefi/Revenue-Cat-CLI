package api

// Project represents a RevenueCat project.
type Project struct {
	Object       string  `json:"object"`
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CreatedAt    int64   `json:"created_at"`
	IconURL      *string `json:"icon_url"`
	IconURLLarge *string `json:"icon_url_large"`
}

// App represents a RevenueCat app.
type App struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	CreatedAt int64  `json:"created_at"`
	ProjectID string `json:"project_id"`
}

// Product represents a RevenueCat product.
type Product struct {
	Object          string               `json:"object"`
	ID              string               `json:"id"`
	StoreIdentifier string               `json:"store_identifier"`
	Type            string               `json:"type"`
	State           string               `json:"state"`
	DisplayName     *string              `json:"display_name"`
	CreatedAt       int64                `json:"created_at"`
	AppID           string               `json:"app_id"`
	App             *App                 `json:"app,omitempty"`
	Subscription    *ProductSubscription `json:"subscription,omitempty"`
	OneTime         *ProductOneTime      `json:"one_time,omitempty"`
}

type ProductSubscription struct {
	Duration            *string `json:"duration"`
	GracePeriodDuration *string `json:"grace_period_duration"`
	TrialDuration       *string `json:"trial_duration"`
}

type ProductOneTime struct {
	IsConsumable *bool `json:"is_consumable"`
}

// Entitlement represents a RevenueCat entitlement.
type Entitlement struct {
	Object      string                 `json:"object"`
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	LookupKey   string                 `json:"lookup_key"`
	DisplayName string                 `json:"display_name"`
	State       string                 `json:"state"`
	CreatedAt   int64                  `json:"created_at"`
	Products    *ListResponse[Product] `json:"products,omitempty"`
}

// Offering represents a RevenueCat offering.
type Offering struct {
	Object      string                 `json:"object"`
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	LookupKey   string                 `json:"lookup_key"`
	DisplayName string                 `json:"display_name"`
	IsCurrent   bool                   `json:"is_current"`
	State       string                 `json:"state"`
	CreatedAt   int64                  `json:"created_at"`
	Metadata    map[string]any         `json:"metadata,omitempty"`
	Packages    *ListResponse[Package] `json:"packages,omitempty"`
}

// Package represents a RevenueCat package within an offering.
type Package struct {
	Object      string                        `json:"object"`
	ID          string                        `json:"id"`
	LookupKey   string                        `json:"lookup_key"`
	DisplayName string                        `json:"display_name"`
	Position    *int                          `json:"position,omitempty"`
	CreatedAt   int64                         `json:"created_at"`
	Products    *ListResponse[PackageProduct] `json:"products,omitempty"`
}

type PackageProduct struct {
	Object              string   `json:"object"`
	ProductID           string   `json:"product_id"`
	EligibilityCriteria string   `json:"eligibility_criteria"`
	Product             *Product `json:"product,omitempty"`
}

// Customer represents a RevenueCat customer/subscriber.
type Customer struct {
	Object                  string                          `json:"object"`
	ID                      string                          `json:"id"`
	ProjectID               string                          `json:"project_id"`
	FirstSeenAt             int64                           `json:"first_seen_at"`
	LastSeenAt              *int64                          `json:"last_seen_at"`
	LastSeenAppVersion      *string                         `json:"last_seen_app_version"`
	LastSeenCountry         *string                         `json:"last_seen_country"`
	LastSeenPlatform        *string                         `json:"last_seen_platform"`
	LastSeenPlatformVersion *string                         `json:"last_seen_platform_version"`
	ActiveEntitlements      ListResponse[ActiveEntitlement] `json:"active_entitlements"`
	Experiment              *ExperimentEnrollment           `json:"experiment,omitempty"`
}

type ActiveEntitlement struct {
	Object        string `json:"object"`
	EntitlementID string `json:"entitlement_id"`
	ExpiresAt     *int64 `json:"expires_at"`
}

type ExperimentEnrollment struct {
	Object  string `json:"object"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Variant string `json:"variant"`
}

// CustomerAlias represents a customer alias.
type CustomerAlias struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// CustomerAttribute represents a customer attribute.
type CustomerAttribute struct {
	Object string `json:"object"`
	Name   string `json:"name"`
	Value  string `json:"value"`
}

// Subscription represents a RevenueCat subscription.
type Subscription struct {
	Object                string              `json:"object"`
	ID                    string              `json:"id"`
	CustomerID            string              `json:"customer_id"`
	OriginalCustomerID    string              `json:"original_customer_id"`
	ProductID             *string             `json:"product_id"`
	StartsAt              int64               `json:"starts_at"`
	CurrentPeriodStartsAt int64               `json:"current_period_starts_at"`
	CurrentPeriodEndsAt   *int64              `json:"current_period_ends_at"`
	EndsAt                *int64              `json:"ends_at"`
	GivesAccess           bool                `json:"gives_access"`
	PendingPayment        bool                `json:"pending_payment"`
	AutoRenewalStatus     string              `json:"auto_renewal_status"`
	Status                string              `json:"status"`
	TotalRevenueInUSD     *MoneyWithBreakdown `json:"total_revenue_in_usd,omitempty"`
	PresentedOfferingID   *string             `json:"presented_offering_id"`
	Environment           string              `json:"environment"`
	Store                 string              `json:"store"`
	StoreSubIdentifier    string              `json:"store_subscription_identifier"`
	Ownership             string              `json:"ownership"`
	Country               *string             `json:"country"`
	ManagementURL         *string             `json:"management_url"`
}

// Transaction represents a subscription transaction.
type Transaction struct {
	Object       string              `json:"object"`
	ID           string              `json:"id"`
	RevenueInUSD *MoneyWithBreakdown `json:"revenue_in_usd,omitempty"`
	PurchasedAt  int64               `json:"purchased_at"`
	Store        string              `json:"store"`
}

type MoneyWithBreakdown struct {
	Currency   string  `json:"currency"`
	Gross      float64 `json:"gross"`
	Commission float64 `json:"commission"`
	Tax        float64 `json:"tax"`
	Proceeds   float64 `json:"proceeds"`
}

// Purchase represents a RevenueCat one-time purchase.
type Purchase struct {
	Object              string              `json:"object"`
	ID                  string              `json:"id"`
	CustomerID          string              `json:"customer_id"`
	OriginalCustomerID  string              `json:"original_customer_id"`
	ProductID           string              `json:"product_id"`
	PurchasedAt         int64               `json:"purchased_at"`
	RevenueInUSD        *MoneyWithBreakdown `json:"revenue_in_usd,omitempty"`
	Quantity            int                 `json:"quantity"`
	Status              string              `json:"status"`
	PresentedOfferingID *string             `json:"presented_offering_id"`
	Environment         string              `json:"environment"`
	Store               string              `json:"store"`
	StorePurchaseID     string              `json:"store_purchase_identifier"`
	Ownership           string              `json:"ownership"`
	Country             *string             `json:"country"`
}

// Webhook represents a webhook integration.
type Webhook struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	CreatedAt int64  `json:"created_at"`
}

// OverviewMetrics represents the metrics overview response.
type OverviewMetrics struct {
	Object  string          `json:"object"`
	Metrics []MetricSummary `json:"metrics"`
}

type MetricSummary struct {
	Object      string  `json:"object"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Description string  `json:"description"`
	Period      string  `json:"period"`
	UpdatedAt   int64   `json:"updated_at"`
}

// ChartData represents chart response.
type ChartData struct {
	Object      string       `json:"object"`
	Name        string       `json:"name"`
	DisplayName string       `json:"display_name"`
	Values      []ChartValue `json:"values"`
}

type ChartValue struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// Paywall represents a RevenueCat paywall.
type Paywall struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

// AuditLogEntry represents an audit log entry.
type AuditLogEntry struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	CreatedAt int64  `json:"created_at"`
	Details   string `json:"details,omitempty"`
}

// Collaborator represents a project collaborator.
type Collaborator struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

// VirtualCurrency represents a virtual currency.
type VirtualCurrency struct {
	Object    string `json:"object"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	State     string `json:"state"`
	CreatedAt int64  `json:"created_at"`
}

// VCBalance represents a customer's virtual currency balance.
type VCBalance struct {
	Object       string `json:"object"`
	CurrencyCode string `json:"currency_code"`
	Balance      int64  `json:"balance"`
}

// VCTransaction represents a virtual currency transaction.
type VCTransaction struct {
	Object       string `json:"object"`
	ID           string `json:"id"`
	CurrencyCode string `json:"currency_code"`
	Amount       int64  `json:"amount"`
	CreatedAt    int64  `json:"created_at"`
}

// Invoice represents a customer invoice.
type Invoice struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
}

// ManagementURL represents an authenticated management URL response.
type ManagementURL struct {
	Object string `json:"object"`
	URL    string `json:"url"`
}

// PublicAPIKey represents a public API key for an app.
type PublicAPIKey struct {
	Object string `json:"object"`
	Key    string `json:"key"`
	Name   string `json:"name"`
}

// ChartOptions represents available options/filters for a chart.
type ChartOptions struct {
	Object  string        `json:"object"`
	Options []ChartOption `json:"options"`
}

type ChartOption struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// DeleteResponse is returned when a resource is deleted.
type DeleteResponse struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	DeletedAt int64  `json:"deleted_at"`
}
