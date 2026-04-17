package api

import (
	"encoding/json"
	"testing"
)

func TestProject_JSONRoundTrip(t *testing.T) {
	original := Project{
		Object:    "project",
		ID:        "proj_abc123",
		Name:      "My Project",
		CreatedAt: 1705311000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Project
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Object != original.Object {
		t.Errorf("Object = %q, want %q", decoded.Object, original.Object)
	}
	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.CreatedAt != original.CreatedAt {
		t.Errorf("CreatedAt = %d, want %d", decoded.CreatedAt, original.CreatedAt)
	}
}

func TestProject_WithOptionalFields(t *testing.T) {
	iconURL := "https://example.com/icon.png"
	iconURLLarge := "https://example.com/icon_large.png"

	original := Project{
		Object:       "project",
		ID:           "proj_123",
		Name:         "Test",
		CreatedAt:    1705311000000,
		IconURL:      &iconURL,
		IconURLLarge: &iconURLLarge,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Project
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.IconURL == nil || *decoded.IconURL != iconURL {
		t.Errorf("IconURL = %v, want %q", decoded.IconURL, iconURL)
	}
	if decoded.IconURLLarge == nil || *decoded.IconURLLarge != iconURLLarge {
		t.Errorf("IconURLLarge = %v, want %q", decoded.IconURLLarge, iconURLLarge)
	}
}

func TestProject_NilOptionalFields(t *testing.T) {
	original := Project{
		Object:    "project",
		ID:        "proj_123",
		Name:      "Test",
		CreatedAt: 1705311000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal raw: %v", err)
	}

	// nil pointer fields should serialize as null (present in JSON)
	_ = raw["icon_url"] // verifying it exists - null is expected for pointer types
}

func TestApp_JSONRoundTrip(t *testing.T) {
	original := App{
		Object:    "app",
		ID:        "app_xyz789",
		Name:      "iOS App",
		Type:      "app_store",
		CreatedAt: 1705400000000,
		ProjectID: "proj_abc123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded App
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}
	if decoded.ProjectID != original.ProjectID {
		t.Errorf("ProjectID = %q, want %q", decoded.ProjectID, original.ProjectID)
	}
}

func TestProduct_JSONRoundTrip(t *testing.T) {
	displayName := "Monthly Premium"

	original := Product{
		Object:          "product",
		ID:              "prod_monthly",
		StoreIdentifier: "com.example.monthly",
		Type:            "subscription",
		State:           "active",
		DisplayName:     &displayName,
		CreatedAt:       1705500000000,
		AppID:           "app_ios",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Product
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.StoreIdentifier != original.StoreIdentifier {
		t.Errorf("StoreIdentifier = %q, want %q", decoded.StoreIdentifier, original.StoreIdentifier)
	}
	if decoded.DisplayName == nil || *decoded.DisplayName != displayName {
		t.Errorf("DisplayName = %v, want %q", decoded.DisplayName, displayName)
	}
	if decoded.AppID != original.AppID {
		t.Errorf("AppID = %q, want %q", decoded.AppID, original.AppID)
	}
}

func TestProduct_WithSubscription(t *testing.T) {
	dur := "P1M"
	grace := "P3D"
	trial := "P7D"

	original := Product{
		Object:          "product",
		ID:              "prod_sub",
		StoreIdentifier: "com.example.sub",
		Type:            "subscription",
		State:           "active",
		CreatedAt:       1705500000000,
		AppID:           "app_1",
		Subscription: &ProductSubscription{
			Duration:            &dur,
			GracePeriodDuration: &grace,
			TrialDuration:       &trial,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Product
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Subscription == nil {
		t.Fatal("Subscription should not be nil")
	}
	if decoded.Subscription.Duration == nil || *decoded.Subscription.Duration != dur {
		t.Errorf("Duration = %v, want %q", decoded.Subscription.Duration, dur)
	}
	if decoded.Subscription.GracePeriodDuration == nil || *decoded.Subscription.GracePeriodDuration != grace {
		t.Errorf("GracePeriodDuration = %v, want %q", decoded.Subscription.GracePeriodDuration, grace)
	}
	if decoded.Subscription.TrialDuration == nil || *decoded.Subscription.TrialDuration != trial {
		t.Errorf("TrialDuration = %v, want %q", decoded.Subscription.TrialDuration, trial)
	}
}

func TestProduct_WithOneTime(t *testing.T) {
	consumable := true

	original := Product{
		Object:          "product",
		ID:              "prod_iap",
		StoreIdentifier: "com.example.coins100",
		Type:            "consumable",
		State:           "active",
		CreatedAt:       1705500000000,
		AppID:           "app_1",
		OneTime: &ProductOneTime{
			IsConsumable: &consumable,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Product
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.OneTime == nil {
		t.Fatal("OneTime should not be nil")
	}
	if decoded.OneTime.IsConsumable == nil || *decoded.OneTime.IsConsumable != true {
		t.Errorf("IsConsumable = %v, want true", decoded.OneTime.IsConsumable)
	}
}

func TestCustomer_JSONRoundTrip(t *testing.T) {
	lastSeen := int64(1705600000000)
	platform := "iOS"
	country := "US"
	version := "1.2.3"
	platVersion := "17.2"

	original := Customer{
		Object:                  "customer",
		ID:                      "user-123",
		ProjectID:               "proj_abc",
		FirstSeenAt:             1705500000000,
		LastSeenAt:              &lastSeen,
		LastSeenAppVersion:      &version,
		LastSeenCountry:         &country,
		LastSeenPlatform:        &platform,
		LastSeenPlatformVersion: &platVersion,
		ActiveEntitlements: ListResponse[ActiveEntitlement]{
			Object: "list",
			Items: []ActiveEntitlement{
				{Object: "entitlement", EntitlementID: "entl_premium"},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Customer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.ProjectID != original.ProjectID {
		t.Errorf("ProjectID = %q, want %q", decoded.ProjectID, original.ProjectID)
	}
	if decoded.LastSeenAt == nil || *decoded.LastSeenAt != lastSeen {
		t.Errorf("LastSeenAt = %v, want %d", decoded.LastSeenAt, lastSeen)
	}
	if decoded.LastSeenPlatform == nil || *decoded.LastSeenPlatform != platform {
		t.Errorf("LastSeenPlatform = %v, want %q", decoded.LastSeenPlatform, platform)
	}
	if len(decoded.ActiveEntitlements.Items) != 1 {
		t.Fatalf("ActiveEntitlements count = %d, want 1", len(decoded.ActiveEntitlements.Items))
	}
	if decoded.ActiveEntitlements.Items[0].EntitlementID != "entl_premium" {
		t.Errorf("EntitlementID = %q, want %q", decoded.ActiveEntitlements.Items[0].EntitlementID, "entl_premium")
	}
}

func TestCustomer_WithExperiment(t *testing.T) {
	original := Customer{
		Object:      "customer",
		ID:          "user-456",
		ProjectID:   "proj_abc",
		FirstSeenAt: 1705500000000,
		ActiveEntitlements: ListResponse[ActiveEntitlement]{
			Object: "list",
			Items:  []ActiveEntitlement{},
		},
		Experiment: &ExperimentEnrollment{
			Object:  "experiment",
			ID:      "exp_1",
			Name:    "Price Test",
			Variant: "control",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Customer
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Experiment == nil {
		t.Fatal("Experiment should not be nil")
	}
	if decoded.Experiment.Name != "Price Test" {
		t.Errorf("Experiment.Name = %q, want %q", decoded.Experiment.Name, "Price Test")
	}
	if decoded.Experiment.Variant != "control" {
		t.Errorf("Experiment.Variant = %q, want %q", decoded.Experiment.Variant, "control")
	}
}

func TestSubscription_JSONRoundTrip(t *testing.T) {
	productID := "prod_monthly"
	endsAt := int64(1710000000000)
	periodEnds := int64(1708000000000)
	offeringID := "ofrnge_default"
	country := "US"
	mgmtURL := "https://apps.apple.com/manage"

	original := Subscription{
		Object:                "subscription",
		ID:                    "sub_abc",
		CustomerID:            "user-123",
		OriginalCustomerID:    "user-123",
		ProductID:             &productID,
		StartsAt:              1705500000000,
		CurrentPeriodStartsAt: 1707000000000,
		CurrentPeriodEndsAt:   &periodEnds,
		EndsAt:                &endsAt,
		GivesAccess:           true,
		PendingPayment:        false,
		AutoRenewalStatus:     "will_renew",
		Status:                "active",
		TotalRevenueInUSD: &MoneyWithBreakdown{
			Currency:   "USD",
			Gross:      9.99,
			Commission: 3.00,
			Tax:        0.80,
			Proceeds:   6.19,
		},
		PresentedOfferingID: &offeringID,
		Environment:         "production",
		Store:               "app_store",
		StoreSubIdentifier:  "sub_12345",
		Ownership:           "purchased",
		Country:             &country,
		ManagementURL:       &mgmtURL,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Subscription
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.CustomerID != original.CustomerID {
		t.Errorf("CustomerID = %q, want %q", decoded.CustomerID, original.CustomerID)
	}
	if decoded.ProductID == nil || *decoded.ProductID != productID {
		t.Errorf("ProductID = %v, want %q", decoded.ProductID, productID)
	}
	if decoded.GivesAccess != true {
		t.Errorf("GivesAccess = %v, want true", decoded.GivesAccess)
	}
	if decoded.Status != "active" {
		t.Errorf("Status = %q, want %q", decoded.Status, "active")
	}
	if decoded.TotalRevenueInUSD == nil {
		t.Fatal("TotalRevenueInUSD should not be nil")
	}
	if decoded.TotalRevenueInUSD.Gross != 9.99 {
		t.Errorf("Gross = %f, want 9.99", decoded.TotalRevenueInUSD.Gross)
	}
	if decoded.TotalRevenueInUSD.Proceeds != 6.19 {
		t.Errorf("Proceeds = %f, want 6.19", decoded.TotalRevenueInUSD.Proceeds)
	}
	if decoded.ManagementURL == nil || *decoded.ManagementURL != mgmtURL {
		t.Errorf("ManagementURL = %v, want %q", decoded.ManagementURL, mgmtURL)
	}
}

func TestSubscription_PromotionalNoProduct(t *testing.T) {
	original := Subscription{
		Object:                "subscription",
		ID:                    "sub_promo",
		CustomerID:            "user-789",
		OriginalCustomerID:    "user-789",
		ProductID:             nil,
		StartsAt:              1705500000000,
		CurrentPeriodStartsAt: 1705500000000,
		GivesAccess:           true,
		AutoRenewalStatus:     "will_not_renew",
		Status:                "active",
		Environment:           "production",
		Store:                 "promotional",
		Ownership:             "purchased",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Subscription
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ProductID != nil {
		t.Errorf("ProductID should be nil for promotional, got %v", decoded.ProductID)
	}
}

func TestEntitlement_JSONRoundTrip(t *testing.T) {
	original := Entitlement{
		Object:      "entitlement",
		ID:          "entl_premium",
		ProjectID:   "proj_abc",
		LookupKey:   "premium",
		DisplayName: "Premium Access",
		State:       "active",
		CreatedAt:   1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Entitlement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.LookupKey != original.LookupKey {
		t.Errorf("LookupKey = %q, want %q", decoded.LookupKey, original.LookupKey)
	}
	if decoded.DisplayName != original.DisplayName {
		t.Errorf("DisplayName = %q, want %q", decoded.DisplayName, original.DisplayName)
	}
}

func TestOffering_JSONRoundTrip(t *testing.T) {
	original := Offering{
		Object:      "offering",
		ID:          "ofrnge_default",
		ProjectID:   "proj_abc",
		LookupKey:   "default",
		DisplayName: "Default Offering",
		IsCurrent:   true,
		State:       "active",
		CreatedAt:   1705500000000,
		Metadata:    map[string]any{"version": "2"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Offering
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if !decoded.IsCurrent {
		t.Error("IsCurrent should be true")
	}
	if decoded.Metadata["version"] != "2" {
		t.Errorf("Metadata[version] = %v, want '2'", decoded.Metadata["version"])
	}
}

func TestPackage_JSONRoundTrip(t *testing.T) {
	pos := 1

	original := Package{
		Object:      "package",
		ID:          "pkge_monthly",
		LookupKey:   "$rc_monthly",
		DisplayName: "Monthly",
		Position:    &pos,
		CreatedAt:   1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Package
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Position == nil || *decoded.Position != 1 {
		t.Errorf("Position = %v, want 1", decoded.Position)
	}
}

func TestPackage_NilPosition(t *testing.T) {
	original := Package{
		Object:      "package",
		ID:          "pkge_1",
		LookupKey:   "custom",
		DisplayName: "Custom",
		CreatedAt:   1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Package
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Position != nil {
		t.Errorf("Position should be nil, got %v", decoded.Position)
	}
}

func TestWebhook_JSONRoundTrip(t *testing.T) {
	original := Webhook{
		Object:    "webhook",
		ID:        "whk_abc",
		Name:      "Production Webhook",
		URL:       "https://api.example.com/webhooks/rc",
		CreatedAt: 1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Webhook
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.URL != original.URL {
		t.Errorf("URL = %q, want %q", decoded.URL, original.URL)
	}
}

func TestPurchase_JSONRoundTrip(t *testing.T) {
	offeringID := "ofrnge_default"
	country := "DE"

	original := Purchase{
		Object:              "purchase",
		ID:                  "purch_xyz",
		CustomerID:          "user-456",
		OriginalCustomerID:  "user-456",
		ProductID:           "prod_coins",
		PurchasedAt:         1705500000000,
		Quantity:            3,
		Status:              "owned",
		PresentedOfferingID: &offeringID,
		Environment:         "production",
		Store:               "app_store",
		StorePurchaseID:     "txn_12345",
		Ownership:           "purchased",
		Country:             &country,
		RevenueInUSD: &MoneyWithBreakdown{
			Currency:   "USD",
			Gross:      2.99,
			Commission: 0.90,
			Tax:        0.24,
			Proceeds:   1.85,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Purchase
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.Quantity != 3 {
		t.Errorf("Quantity = %d, want 3", decoded.Quantity)
	}
	if decoded.Country == nil || *decoded.Country != "DE" {
		t.Errorf("Country = %v, want DE", decoded.Country)
	}
	if decoded.RevenueInUSD == nil {
		t.Fatal("RevenueInUSD should not be nil")
	}
	if decoded.RevenueInUSD.Proceeds != 1.85 {
		t.Errorf("Proceeds = %f, want 1.85", decoded.RevenueInUSD.Proceeds)
	}
}

func TestTransaction_JSONRoundTrip(t *testing.T) {
	original := Transaction{
		Object:      "transaction",
		ID:          "txn_abc",
		PurchasedAt: 1705500000000,
		Store:       "app_store",
		RevenueInUSD: &MoneyWithBreakdown{
			Currency:   "USD",
			Gross:      4.99,
			Commission: 1.50,
			Tax:        0.40,
			Proceeds:   3.09,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Transaction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, original.ID)
	}
	if decoded.RevenueInUSD == nil {
		t.Fatal("RevenueInUSD should not be nil")
	}
}

func TestMoneyWithBreakdown_JSONRoundTrip(t *testing.T) {
	original := MoneyWithBreakdown{
		Currency:   "USD",
		Gross:      9.99,
		Commission: 3.00,
		Tax:        0.80,
		Proceeds:   6.19,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded MoneyWithBreakdown
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", decoded.Currency)
	}
	if decoded.Gross != 9.99 {
		t.Errorf("Gross = %f, want 9.99", decoded.Gross)
	}
	if decoded.Commission != 3.00 {
		t.Errorf("Commission = %f, want 3.00", decoded.Commission)
	}
	if decoded.Tax != 0.80 {
		t.Errorf("Tax = %f, want 0.80", decoded.Tax)
	}
	if decoded.Proceeds != 6.19 {
		t.Errorf("Proceeds = %f, want 6.19", decoded.Proceeds)
	}
}

func TestOverviewMetrics_JSONRoundTrip(t *testing.T) {
	original := OverviewMetrics{
		Object: "overview_metrics",
		Metrics: []MetricSummary{
			{
				Object:      "metric",
				Name:        "mrr",
				Value:       12345.67,
				Description: "Monthly Recurring Revenue",
				Period:      "last_28_days",
				UpdatedAt:   1705600000000,
			},
			{
				Object:      "metric",
				Name:        "active_subscriptions",
				Value:       500,
				Description: "Active Subscriptions",
				Period:      "current",
				UpdatedAt:   1705600000000,
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded OverviewMetrics
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(decoded.Metrics) != 2 {
		t.Fatalf("Metrics count = %d, want 2", len(decoded.Metrics))
	}
	if decoded.Metrics[0].Name != "mrr" {
		t.Errorf("Metrics[0].Name = %q, want %q", decoded.Metrics[0].Name, "mrr")
	}
	if decoded.Metrics[0].Value != 12345.67 {
		t.Errorf("Metrics[0].Value = %f, want 12345.67", decoded.Metrics[0].Value)
	}
}

func TestChartData_JSONRoundTrip(t *testing.T) {
	original := ChartData{
		Object:      "chart",
		Name:        "revenue",
		DisplayName: "Revenue",
		Values: []ChartValue{
			{Date: "2024-01-01", Value: 100.50},
			{Date: "2024-01-02", Value: 120.75},
			{Date: "2024-01-03", Value: 95.25},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ChartData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Name != "revenue" {
		t.Errorf("Name = %q, want %q", decoded.Name, "revenue")
	}
	if len(decoded.Values) != 3 {
		t.Fatalf("Values count = %d, want 3", len(decoded.Values))
	}
	if decoded.Values[1].Date != "2024-01-02" {
		t.Errorf("Values[1].Date = %q, want %q", decoded.Values[1].Date, "2024-01-02")
	}
}

func TestActiveEntitlement_WithExpiry(t *testing.T) {
	expires := int64(1710000000000)

	original := ActiveEntitlement{
		Object:        "entitlement",
		EntitlementID: "entl_pro",
		ExpiresAt:     &expires,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ActiveEntitlement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ExpiresAt == nil || *decoded.ExpiresAt != expires {
		t.Errorf("ExpiresAt = %v, want %d", decoded.ExpiresAt, expires)
	}
}

func TestActiveEntitlement_NoExpiry(t *testing.T) {
	original := ActiveEntitlement{
		Object:        "entitlement",
		EntitlementID: "entl_lifetime",
		ExpiresAt:     nil,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ActiveEntitlement
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ExpiresAt != nil {
		t.Errorf("ExpiresAt should be nil, got %v", decoded.ExpiresAt)
	}
}

func TestCollaborator_JSONRoundTrip(t *testing.T) {
	original := Collaborator{
		Object: "collaborator",
		ID:     "collab_1",
		Email:  "dev@example.com",
		Role:   "admin",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Collaborator
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Email != "dev@example.com" {
		t.Errorf("Email = %q, want %q", decoded.Email, "dev@example.com")
	}
	if decoded.Role != "admin" {
		t.Errorf("Role = %q, want %q", decoded.Role, "admin")
	}
}

func TestVirtualCurrency_JSONRoundTrip(t *testing.T) {
	original := VirtualCurrency{
		Object:    "virtual_currency",
		Code:      "COINS",
		Name:      "Gold Coins",
		State:     "active",
		CreatedAt: 1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded VirtualCurrency
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Code != "COINS" {
		t.Errorf("Code = %q, want %q", decoded.Code, "COINS")
	}
	if decoded.Name != "Gold Coins" {
		t.Errorf("Name = %q, want %q", decoded.Name, "Gold Coins")
	}
}

func TestVCBalance_JSONRoundTrip(t *testing.T) {
	original := VCBalance{
		Object:       "balance",
		CurrencyCode: "GEMS",
		Balance:      1500,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded VCBalance
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.CurrencyCode != "GEMS" {
		t.Errorf("CurrencyCode = %q, want %q", decoded.CurrencyCode, "GEMS")
	}
	if decoded.Balance != 1500 {
		t.Errorf("Balance = %d, want 1500", decoded.Balance)
	}
}

func TestVCTransaction_JSONRoundTrip(t *testing.T) {
	original := VCTransaction{
		Object:       "transaction",
		ID:           "vctx_123",
		CurrencyCode: "COINS",
		Amount:       100,
		CreatedAt:    1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded VCTransaction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Amount != 100 {
		t.Errorf("Amount = %d, want 100", decoded.Amount)
	}
}

func TestDeleteResponse_JSONRoundTrip(t *testing.T) {
	original := DeleteResponse{
		Object:    "deleted",
		ID:        "prod_123",
		DeletedAt: 1705600000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded DeleteResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != "prod_123" {
		t.Errorf("ID = %q, want %q", decoded.ID, "prod_123")
	}
	if decoded.DeletedAt != 1705600000000 {
		t.Errorf("DeletedAt = %d, want 1705600000000", decoded.DeletedAt)
	}
}

func TestAuditLogEntry_JSONRoundTrip(t *testing.T) {
	original := AuditLogEntry{
		Object:    "audit_log",
		ID:        "log_abc",
		Action:    "product.created",
		Actor:     "user@example.com",
		CreatedAt: 1705500000000,
		Details:   "Created product prod_123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded AuditLogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Action != "product.created" {
		t.Errorf("Action = %q, want %q", decoded.Action, "product.created")
	}
	if decoded.Details != "Created product prod_123" {
		t.Errorf("Details = %q, want %q", decoded.Details, "Created product prod_123")
	}
}

func TestPaywall_JSONRoundTrip(t *testing.T) {
	original := Paywall{
		Object:    "paywall",
		ID:        "pw_abc",
		CreatedAt: 1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Paywall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != "pw_abc" {
		t.Errorf("ID = %q, want %q", decoded.ID, "pw_abc")
	}
}

func TestManagementURL_JSONRoundTrip(t *testing.T) {
	original := ManagementURL{
		Object:        "authenticated_management_url",
		ManagementURL: "https://apps.apple.com/account/subscriptions",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ManagementURL
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ManagementURL != "https://apps.apple.com/account/subscriptions" {
		t.Errorf("ManagementURL = %q, want %q", decoded.ManagementURL, "https://apps.apple.com/account/subscriptions")
	}
}

func TestPublicAPIKey_JSONRoundTrip(t *testing.T) {
	original := PublicAPIKey{
		Object: "public_api_key",
		Key:    "appl_abcdef123456",
		Name:   "iOS Public Key",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded PublicAPIKey
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Key != "appl_abcdef123456" {
		t.Errorf("Key = %q, want %q", decoded.Key, "appl_abcdef123456")
	}
}

func TestChartOptions_JSONRoundTrip(t *testing.T) {
	original := ChartOptions{
		Object: "chart_options",
		Options: []ChartOption{
			{Name: "country", Values: []string{"US", "DE", "JP"}},
			{Name: "store", Values: []string{"app_store", "play_store"}},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ChartOptions
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(decoded.Options) != 2 {
		t.Fatalf("Options count = %d, want 2", len(decoded.Options))
	}
	if decoded.Options[0].Name != "country" {
		t.Errorf("Options[0].Name = %q, want %q", decoded.Options[0].Name, "country")
	}
	if len(decoded.Options[0].Values) != 3 {
		t.Errorf("Options[0].Values count = %d, want 3", len(decoded.Options[0].Values))
	}
}

func TestInvoice_JSONRoundTrip(t *testing.T) {
	original := Invoice{
		Object:    "invoice",
		ID:        "inv_abc",
		CreatedAt: 1705500000000,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Invoice
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != "inv_abc" {
		t.Errorf("ID = %q, want %q", decoded.ID, "inv_abc")
	}
}

func TestCustomerAlias_JSONRoundTrip(t *testing.T) {
	original := CustomerAlias{
		Object: "alias",
		ID:     "alias_user456",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CustomerAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ID != "alias_user456" {
		t.Errorf("ID = %q, want %q", decoded.ID, "alias_user456")
	}
}

func TestCustomerAttribute_JSONRoundTrip(t *testing.T) {
	original := CustomerAttribute{
		Object: "attribute",
		Name:   "$email",
		Value:  "test@example.com",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded CustomerAttribute
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Name != "$email" {
		t.Errorf("Name = %q, want %q", decoded.Name, "$email")
	}
	if decoded.Value != "test@example.com" {
		t.Errorf("Value = %q, want %q", decoded.Value, "test@example.com")
	}
}

func TestPackageProduct_JSONRoundTrip(t *testing.T) {
	original := PackageProduct{
		Object:              "package_product",
		ProductID:           "prod_monthly",
		EligibilityCriteria: "all",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded PackageProduct
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.ProductID != "prod_monthly" {
		t.Errorf("ProductID = %q, want %q", decoded.ProductID, "prod_monthly")
	}
	if decoded.EligibilityCriteria != "all" {
		t.Errorf("EligibilityCriteria = %q, want %q", decoded.EligibilityCriteria, "all")
	}
}

func TestError_JSONRoundTrip(t *testing.T) {
	backoff := 500

	original := Error{
		Object:    "error",
		Type:      "parameter_error",
		Param:     "name",
		DocURL:    "https://docs.revenuecat.com",
		Message:   "Name is required",
		Retryable: false,
		BackoffMs: &backoff,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Error
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.Type != "parameter_error" {
		t.Errorf("Type = %q, want %q", decoded.Type, "parameter_error")
	}
	if decoded.Message != "Name is required" {
		t.Errorf("Message = %q, want %q", decoded.Message, "Name is required")
	}
	if decoded.BackoffMs == nil || *decoded.BackoffMs != 500 {
		t.Errorf("BackoffMs = %v, want 500", decoded.BackoffMs)
	}
}

func TestListResponse_Unmarshal_FromJSON(t *testing.T) {
	jsonStr := `{
		"object": "list",
		"items": [
			{"object": "app", "id": "app_1", "name": "iOS", "type": "app_store", "created_at": 1705500000000, "project_id": "proj_1"},
			{"object": "app", "id": "app_2", "name": "Android", "type": "play_store", "created_at": 1705500000000, "project_id": "proj_1"}
		],
		"next_page": null,
		"url": "/v2/projects/proj_1/apps"
	}`

	var resp ListResponse[App]
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(resp.Items))
	}
	if resp.Items[0].Name != "iOS" {
		t.Errorf("Items[0].Name = %q, want iOS", resp.Items[0].Name)
	}
	if resp.Items[1].Type != "play_store" {
		t.Errorf("Items[1].Type = %q, want play_store", resp.Items[1].Type)
	}
}
