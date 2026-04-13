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

// Product represents a RevenueCat product.
type Product struct {
	Object          string              `json:"object"`
	ID              string              `json:"id"`
	StoreIdentifier string              `json:"store_identifier"`
	Type            string              `json:"type"`
	State           string              `json:"state"`
	DisplayName     *string             `json:"display_name"`
	CreatedAt       int64               `json:"created_at"`
	AppID           string              `json:"app_id"`
	App             *App                `json:"app,omitempty"`
	Subscription    *ProductSubscription `json:"subscription,omitempty"`
	OneTime         *ProductOneTime     `json:"one_time,omitempty"`
}

type ProductSubscription struct {
	Duration            *string `json:"duration"`
	GracePeriodDuration *string `json:"grace_period_duration"`
	TrialDuration       *string `json:"trial_duration"`
}

type ProductOneTime struct {
	IsConsumable *bool `json:"is_consumable"`
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

// Entitlement represents a RevenueCat entitlement.
type Entitlement struct {
	Object    string  `json:"object"`
	ID        string  `json:"id"`
	ProjectID string  `json:"project_id"`
	LookupKey string  `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	State     string  `json:"state"`
	CreatedAt int64   `json:"created_at"`
	Products  *ListResponse[Product] `json:"products,omitempty"`
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
	Object      string `json:"object"`
	ID          string `json:"id"`
	LookupKey   string `json:"lookup_key"`
	DisplayName string `json:"display_name"`
	Position    *int   `json:"position,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	Products    *ListResponse[PackageProduct] `json:"products,omitempty"`
}

type PackageProduct struct {
	Object             string  `json:"object"`
	ProductID          string  `json:"product_id"`
	EligibilityCriteria string `json:"eligibility_criteria"`
	Product            *Product `json:"product,omitempty"`
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

// Subscription represents a RevenueCat subscription.
type Subscription struct {
	Object                  string          `json:"object"`
	ID                      string          `json:"id"`
	CustomerID              string          `json:"customer_id"`
	ProductID               *string         `json:"product_id"`
	StartsAt                int64           `json:"starts_at"`
	CurrentPeriodStartsAt   int64           `json:"current_period_starts_at"`
	CurrentPeriodEndsAt     *int64          `json:"current_period_ends_at"`
	GivesAccess             bool            `json:"gives_access"`
	AutoRenewalStatus       string          `json:"auto_renewal_status"`
	Status                  string          `json:"status"`
	TotalRevenueInUSD       *MoneyWithBreakdown `json:"total_revenue_in_usd,omitempty"`
	Environment             string          `json:"environment"`
	Store                   string          `json:"store"`
}

type MoneyWithBreakdown struct {
	Currency   string  `json:"currency"`
	Gross      float64 `json:"gross"`
	Commission float64 `json:"commission"`
	Tax        float64 `json:"tax"`
	Proceeds   float64 `json:"proceeds"`
}

// DeleteResponse is returned when a resource is deleted.
type DeleteResponse struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	DeletedAt int64  `json:"deleted_at"`
}
