package customerdiagnosis

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/andresdefi/rc/internal/api"
)

const (
	StatusPass = "pass"
	StatusWarn = "warn"
	StatusFail = "fail"
	StatusInfo = "info"

	AccessHasAccess = "has_access"
	AccessNoAccess  = "no_access"
)

type Report struct {
	Object             string                `json:"object"`
	ProjectID          string                `json:"project_id"`
	CustomerID         string                `json:"customer_id"`
	AccessSummary      string                `json:"access_summary"`
	Status             string                `json:"status"`
	Counts             Counts                `json:"counts"`
	ActiveEntitlements []EntitlementSummary  `json:"active_entitlements"`
	Subscriptions      []SubscriptionSummary `json:"subscriptions"`
	Purchases          []PurchaseSummary     `json:"purchases"`
	Aliases            []string              `json:"aliases"`
	Attributes         []AttributeSummary    `json:"attributes,omitempty"`
	Findings           []Finding             `json:"findings"`
	NextCommands       []string              `json:"next_commands"`
}

type Counts struct {
	ActiveEntitlements   int `json:"active_entitlements"`
	Subscriptions        int `json:"subscriptions"`
	ActiveSubscriptions  int `json:"active_subscriptions"`
	ExpiredSubscriptions int `json:"expired_subscriptions"`
	Purchases            int `json:"purchases"`
	ActivePurchases      int `json:"active_purchases"`
	Aliases              int `json:"aliases"`
	Attributes           int `json:"attributes"`
}

type EntitlementSummary struct {
	EntitlementID string `json:"entitlement_id"`
	ExpiresAt     *int64 `json:"expires_at"`
}

type SubscriptionSummary struct {
	ID                  string `json:"id"`
	ProductID           string `json:"product_id"`
	Status              string `json:"status"`
	Store               string `json:"store"`
	Environment         string `json:"environment"`
	GivesAccess         bool   `json:"gives_access"`
	PendingPayment      bool   `json:"pending_payment"`
	AutoRenewalStatus   string `json:"auto_renewal_status"`
	CurrentPeriodEndsAt *int64 `json:"current_period_ends_at"`
	EndsAt              *int64 `json:"ends_at"`
}

type PurchaseSummary struct {
	ID          string `json:"id"`
	ProductID   string `json:"product_id"`
	Status      string `json:"status"`
	Store       string `json:"store"`
	Environment string `json:"environment"`
	Quantity    int    `json:"quantity"`
	PurchasedAt int64  `json:"purchased_at"`
}

type AttributeSummary struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Finding struct {
	Severity string   `json:"severity"`
	Area     string   `json:"area"`
	Message  string   `json:"message"`
	Details  []string `json:"details,omitempty"`
}

func Analyze(client *api.Client, projectID, customerID string) (*Report, error) {
	escapedProjectID := url.PathEscape(projectID)
	escapedCustomerID := url.PathEscape(customerID)

	customer, err := fetchCustomer(client, escapedProjectID, escapedCustomerID, customerID)
	if err != nil {
		return nil, err
	}

	activeEntitlements, err := fetchList[api.ActiveEntitlement](client, escapedProjectID, escapedCustomerID, "active_entitlements", "active entitlements", customerID)
	if err != nil {
		return nil, err
	}
	subscriptions, err := fetchList[api.Subscription](client, escapedProjectID, escapedCustomerID, "subscriptions", "subscriptions", customerID)
	if err != nil {
		return nil, err
	}
	purchases, err := fetchList[api.Purchase](client, escapedProjectID, escapedCustomerID, "purchases", "purchases", customerID)
	if err != nil {
		return nil, err
	}
	aliases, err := fetchList[api.CustomerAlias](client, escapedProjectID, escapedCustomerID, "aliases", "aliases", customerID)
	if err != nil {
		return nil, err
	}
	attributes, err := fetchList[api.CustomerAttribute](client, escapedProjectID, escapedCustomerID, "attributes", "attributes", customerID)
	if err != nil {
		return nil, err
	}

	return buildReport(projectID, customerID, customer, activeEntitlements, subscriptions, purchases, aliases, attributes), nil
}

func fetchCustomer(client *api.Client, escapedProjectID, escapedCustomerID, customerID string) (api.Customer, error) {
	data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s", escapedProjectID, escapedCustomerID), nil)
	if err != nil {
		return api.Customer{}, fmt.Errorf("lookup customer %s: %w", customerID, err)
	}
	var customer api.Customer
	if err := json.Unmarshal(data, &customer); err != nil {
		return api.Customer{}, fmt.Errorf("parse customer %s: %w", customerID, err)
	}
	return customer, nil
}

func fetchList[T any](client *api.Client, escapedProjectID, escapedCustomerID, resource, label, customerID string) ([]T, error) {
	path := fmt.Sprintf("/projects/%s/customers/%s/%s", escapedProjectID, escapedCustomerID, resource)
	items, err := api.PaginateAll[T](client, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list customer %s for %s: %w", label, customerID, err)
	}
	return items, nil
}

func buildReport(
	projectID string,
	customerID string,
	customer api.Customer,
	activeEntitlements []api.ActiveEntitlement,
	subscriptions []api.Subscription,
	purchases []api.Purchase,
	aliases []api.CustomerAlias,
	attributes []api.CustomerAttribute,
) *Report {
	if len(activeEntitlements) == 0 && len(customer.ActiveEntitlements.Items) > 0 {
		activeEntitlements = customer.ActiveEntitlements.Items
	}

	report := &Report{
		Object:             "customer_diagnosis",
		ProjectID:          projectID,
		CustomerID:         customerID,
		AccessSummary:      AccessNoAccess,
		Status:             StatusPass,
		ActiveEntitlements: summarizeEntitlements(activeEntitlements),
		Subscriptions:      summarizeSubscriptions(subscriptions),
		Purchases:          summarizePurchases(purchases),
		Aliases:            summarizeAliases(aliases),
		Attributes:         summarizeAttributes(attributes),
		NextCommands:       nextCommands(customerID, activeEntitlements),
	}
	report.Counts.ActiveEntitlements = len(activeEntitlements)
	report.Counts.Subscriptions = len(subscriptions)
	report.Counts.Purchases = len(purchases)
	report.Counts.Aliases = len(aliases)
	report.Counts.Attributes = len(attributes)

	if len(activeEntitlements) > 0 {
		report.AccessSummary = AccessHasAccess
	}

	report.add(Finding{
		Severity: StatusPass,
		Area:     "customer",
		Message:  fmt.Sprintf("Customer %s exists in project %s.", customer.ID, projectID),
	})
	report.checkEntitlements(activeEntitlements)
	report.checkSubscriptions(subscriptions, len(activeEntitlements) > 0)
	report.checkPurchases(purchases, len(activeEntitlements) > 0)
	report.checkAliases(aliases)
	report.checkAttributes(attributes)

	return report
}

func summarizeEntitlements(items []api.ActiveEntitlement) []EntitlementSummary {
	result := make([]EntitlementSummary, 0, len(items))
	for _, item := range items {
		result = append(result, EntitlementSummary{
			EntitlementID: item.EntitlementID,
			ExpiresAt:     item.ExpiresAt,
		})
	}
	return result
}

func summarizeSubscriptions(items []api.Subscription) []SubscriptionSummary {
	result := make([]SubscriptionSummary, 0, len(items))
	for _, item := range items {
		result = append(result, SubscriptionSummary{
			ID:                  item.ID,
			ProductID:           deref(item.ProductID, "promotional"),
			Status:              item.Status,
			Store:               item.Store,
			Environment:         item.Environment,
			GivesAccess:         item.GivesAccess,
			PendingPayment:      item.PendingPayment,
			AutoRenewalStatus:   item.AutoRenewalStatus,
			CurrentPeriodEndsAt: item.CurrentPeriodEndsAt,
			EndsAt:              item.EndsAt,
		})
	}
	return result
}

func summarizePurchases(items []api.Purchase) []PurchaseSummary {
	result := make([]PurchaseSummary, 0, len(items))
	for _, item := range items {
		result = append(result, PurchaseSummary{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Status:      item.Status,
			Store:       item.Store,
			Environment: item.Environment,
			Quantity:    item.Quantity,
			PurchasedAt: item.PurchasedAt,
		})
	}
	return result
}

func summarizeAliases(items []api.CustomerAlias) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.ID)
	}
	return result
}

func summarizeAttributes(items []api.CustomerAttribute) []AttributeSummary {
	result := make([]AttributeSummary, 0, len(items))
	for _, item := range items {
		result = append(result, AttributeSummary{Name: item.Name, Value: item.Value})
	}
	return result
}

func nextCommands(customerID string, activeEntitlements []api.ActiveEntitlement) []string {
	commands := []string{
		fmt.Sprintf("rc customers entitlements %s", customerID),
		fmt.Sprintf("rc customers subscriptions %s --all", customerID),
		fmt.Sprintf("rc customers purchases %s --all", customerID),
		fmt.Sprintf("rc customers aliases %s", customerID),
	}
	seen := map[string]bool{}
	for _, entitlement := range activeEntitlements {
		if entitlement.EntitlementID == "" || seen[entitlement.EntitlementID] {
			continue
		}
		commands = append(commands, fmt.Sprintf("rc entitlements products %s --all", entitlement.EntitlementID))
		seen[entitlement.EntitlementID] = true
	}
	return commands
}

func (r *Report) add(f Finding) {
	r.Findings = append(r.Findings, f)
	switch f.Severity {
	case StatusFail:
		r.Status = StatusFail
	case StatusWarn:
		if r.Status != StatusFail {
			r.Status = StatusWarn
		}
	}
}

func (r *Report) checkEntitlements(activeEntitlements []api.ActiveEntitlement) {
	if len(activeEntitlements) == 0 {
		r.add(Finding{
			Severity: StatusFail,
			Area:     "access",
			Message:  "No active entitlements are present, so RevenueCat is not granting entitlement-based access.",
			Details:  []string{fmt.Sprintf("Run rc customers entitlements %s to confirm the active entitlement list.", r.CustomerID)},
		})
		return
	}

	var details []string
	for _, entitlement := range activeEntitlements {
		details = append(details, entitlement.EntitlementID)
	}
	r.add(Finding{
		Severity: StatusPass,
		Area:     "access",
		Message:  fmt.Sprintf("%d active entitlement(s) grant access.", len(activeEntitlements)),
		Details:  details,
	})
}

func (r *Report) checkSubscriptions(subscriptions []api.Subscription, hasEntitlement bool) {
	if len(subscriptions) == 0 {
		r.add(Finding{Severity: StatusInfo, Area: "subscriptions", Message: "No subscriptions found for this customer."})
		return
	}

	var (
		activeNoAccessDetails []string
		noAccessDetails       []string
		expiredDetails        []string
		stateDetails          []string
		unknownDetails        []string
	)

	for _, subscription := range subscriptions {
		status := normalizeStatus(subscription.Status)
		if isActiveSubscriptionStatus(status) {
			r.Counts.ActiveSubscriptions++
			if !subscription.GivesAccess || !hasEntitlement {
				activeNoAccessDetails = append(activeNoAccessDetails, subscriptionLabel(subscription))
			}
		}
		if !subscription.GivesAccess {
			noAccessDetails = append(noAccessDetails, subscriptionLabel(subscription))
		}
		if isExpiredSubscriptionStatus(status) {
			r.Counts.ExpiredSubscriptions++
			expiredDetails = append(expiredDetails, subscriptionLabel(subscription))
		}
		if isAttentionSubscriptionStatus(status) || subscription.PendingPayment {
			stateDetails = append(stateDetails, subscriptionLabel(subscription))
		}
		if !isKnownSubscriptionStatus(status) {
			unknownDetails = append(unknownDetails, subscriptionLabel(subscription))
		}
	}

	if len(activeNoAccessDetails) > 0 {
		r.add(Finding{
			Severity: StatusFail,
			Area:     "subscriptions",
			Message:  "Active subscription(s) are present but do not line up with active entitlement access.",
			Details:  activeNoAccessDetails,
		})
	} else if r.Counts.ActiveSubscriptions > 0 {
		r.add(Finding{
			Severity: StatusPass,
			Area:     "subscriptions",
			Message:  fmt.Sprintf("%d active subscription(s) are present.", r.Counts.ActiveSubscriptions),
		})
	}

	if len(noAccessDetails) > 0 {
		r.add(Finding{
			Severity: StatusWarn,
			Area:     "subscriptions",
			Message:  "Some subscription products currently report gives_access=false.",
			Details:  noAccessDetails,
		})
	}
	if len(expiredDetails) > 0 {
		r.add(Finding{
			Severity: StatusInfo,
			Area:     "subscriptions",
			Message:  "Expired subscription(s) are present in purchase history.",
			Details:  expiredDetails,
		})
	}
	if len(stateDetails) > 0 {
		r.add(Finding{
			Severity: StatusWarn,
			Area:     "subscriptions",
			Message:  "Subscription state needs attention, such as billing retry, grace period, paused access, or pending payment.",
			Details:  stateDetails,
		})
	}
	if len(unknownDetails) > 0 {
		r.add(Finding{
			Severity: StatusInfo,
			Area:     "subscriptions",
			Message:  "Subscription status is not recognized by this CLI; inspect the subscription directly.",
			Details:  unknownDetails,
		})
	}
}

func (r *Report) checkPurchases(purchases []api.Purchase, hasEntitlement bool) {
	if len(purchases) == 0 {
		if r.Counts.Subscriptions == 0 {
			r.add(Finding{
				Severity: StatusFail,
				Area:     "purchases",
				Message:  "No purchases or subscriptions were found for this customer.",
			})
			return
		}
		r.add(Finding{Severity: StatusInfo, Area: "purchases", Message: "No one-time purchases found for this customer."})
		return
	}

	var activeDetails []string
	for _, purchase := range purchases {
		if isActivePurchaseStatus(normalizeStatus(purchase.Status)) {
			r.Counts.ActivePurchases++
			activeDetails = append(activeDetails, purchaseLabel(purchase))
		}
	}

	if r.Counts.ActivePurchases > 0 && !hasEntitlement {
		r.add(Finding{
			Severity: StatusFail,
			Area:     "purchases",
			Message:  "Active purchase(s) exist, but the customer has no active entitlement access.",
			Details:  activeDetails,
		})
		return
	}
	if r.Counts.ActivePurchases > 0 {
		r.add(Finding{
			Severity: StatusPass,
			Area:     "purchases",
			Message:  fmt.Sprintf("%d active purchase(s) are present.", r.Counts.ActivePurchases),
			Details:  activeDetails,
		})
		return
	}
	r.add(Finding{Severity: StatusInfo, Area: "purchases", Message: "Purchase history exists, but no active purchase status was found."})
}

func (r *Report) checkAliases(aliases []api.CustomerAlias) {
	if len(aliases) == 0 {
		r.add(Finding{Severity: StatusInfo, Area: "aliases", Message: "No aliases found for this customer."})
		return
	}
	r.add(Finding{
		Severity: StatusInfo,
		Area:     "aliases",
		Message:  "Aliases are present and may explain split identity or transferred purchases.",
		Details:  summarizeAliases(aliases),
	})
}

func (r *Report) checkAttributes(attributes []api.CustomerAttribute) {
	if len(attributes) == 0 {
		return
	}
	names := make([]string, 0, len(attributes))
	for _, attribute := range attributes {
		names = append(names, attribute.Name)
	}
	r.add(Finding{
		Severity: StatusInfo,
		Area:     "attributes",
		Message:  fmt.Sprintf("%d customer attribute(s) found.", len(attributes)),
		Details:  names,
	})
}

func normalizeStatus(status string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(status)), "-", "_")
}

func isActiveSubscriptionStatus(status string) bool {
	switch status {
	case "active", "trialing", "intro", "in_intro_offer", "in_trial":
		return true
	default:
		return false
	}
}

func isExpiredSubscriptionStatus(status string) bool {
	switch status {
	case "expired", "cancelled", "canceled", "revoked":
		return true
	default:
		return false
	}
}

func isAttentionSubscriptionStatus(status string) bool {
	switch status {
	case "billing_retry", "billing_issue", "in_grace_period", "grace_period", "paused":
		return true
	default:
		return false
	}
}

func isKnownSubscriptionStatus(status string) bool {
	return isActiveSubscriptionStatus(status) || isExpiredSubscriptionStatus(status) || isAttentionSubscriptionStatus(status) || status == "refunded"
}

func isActivePurchaseStatus(status string) bool {
	switch status {
	case "active", "purchased":
		return true
	default:
		return false
	}
}

func subscriptionLabel(subscription api.Subscription) string {
	return fmt.Sprintf("%s product=%s status=%s store=%s gives_access=%t", subscription.ID, deref(subscription.ProductID, "promotional"), subscription.Status, subscription.Store, subscription.GivesAccess)
}

func purchaseLabel(purchase api.Purchase) string {
	return fmt.Sprintf("%s product=%s status=%s store=%s", purchase.ID, purchase.ProductID, purchase.Status, purchase.Store)
}

func deref(value *string, fallback string) string {
	if value == nil || *value == "" {
		return fallback
	}
	return *value
}
