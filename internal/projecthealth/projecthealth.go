package projecthealth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/andresdefi/rc/internal/api"
)

const (
	StatusPass = "pass"
	StatusWarn = "warn"
	StatusFail = "fail"
)

type Report struct {
	Object    string  `json:"object"`
	ProjectID string  `json:"project_id"`
	Status    string  `json:"status"`
	Counts    Counts  `json:"counts"`
	Checks    []Check `json:"checks"`
}

type Counts struct {
	Apps                    int `json:"apps"`
	Products                int `json:"products"`
	ActiveProducts          int `json:"active_products"`
	Entitlements            int `json:"entitlements"`
	ActiveEntitlements      int `json:"active_entitlements"`
	EntitlementProductLinks int `json:"entitlement_product_links"`
	Offerings               int `json:"offerings"`
	ActiveOfferings         int `json:"active_offerings"`
	CurrentOfferings        int `json:"current_offerings"`
	Packages                int `json:"packages"`
	PackageProductLinks     int `json:"package_product_links"`
}

type Check struct {
	Status  string   `json:"status"`
	Area    string   `json:"area"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type LaunchReport struct {
	Object        string  `json:"object"`
	ProjectID     string  `json:"project_id"`
	Ready         bool    `json:"ready"`
	Status        string  `json:"status"`
	ProjectStatus string  `json:"project_status"`
	Counts        Counts  `json:"counts"`
	Checks        []Check `json:"checks"`
}

type entitlementProducts struct {
	entitlement api.Entitlement
	products    []api.Product
}

type offeringPackages struct {
	offering api.Offering
	packages []packageProducts
}

type packageProducts struct {
	pkg      api.Package
	products []api.PackageProduct
}

func Analyze(client *api.Client, projectID string) (*Report, error) {
	snapshot, err := fetchSnapshot(client, projectID)
	if err != nil {
		return nil, err
	}
	return buildReport(projectID, snapshot), nil
}

func AssessLaunch(report *Report) *LaunchReport {
	launch := &LaunchReport{
		Object:        "launch_check_report",
		ProjectID:     report.ProjectID,
		Ready:         true,
		Status:        StatusPass,
		ProjectStatus: report.Status,
		Counts:        report.Counts,
	}

	launch.add(checkProjectHealth(report))
	launch.add(checkLaunchCatalog(report.Counts))
	launch.add(checkLaunchEntitlements(report.Counts))
	launch.add(checkLaunchOffering(report.Counts))
	launch.add(checkLaunchPackages(report.Counts))

	return launch
}

type snapshot struct {
	apps         []api.App
	products     []api.Product
	entitlements []entitlementProducts
	offerings    []offeringPackages
}

func fetchSnapshot(client *api.Client, projectID string) (*snapshot, error) {
	escapedProjectID := url.PathEscape(projectID)

	apps, err := api.PaginateAll[api.App](client, fmt.Sprintf("/projects/%s/apps", escapedProjectID), nil)
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}
	products, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/products", escapedProjectID), nil)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	entitlements, err := fetchEntitlements(client, escapedProjectID)
	if err != nil {
		return nil, err
	}
	offerings, err := fetchOfferings(client, escapedProjectID)
	if err != nil {
		return nil, err
	}

	return &snapshot{
		apps:         apps,
		products:     products,
		entitlements: entitlements,
		offerings:    offerings,
	}, nil
}

func fetchEntitlements(client *api.Client, escapedProjectID string) ([]entitlementProducts, error) {
	items, err := api.PaginateAll[api.Entitlement](client, fmt.Sprintf("/projects/%s/entitlements", escapedProjectID), nil)
	if err != nil {
		return nil, fmt.Errorf("list entitlements: %w", err)
	}
	result := make([]entitlementProducts, 0, len(items))
	for _, entitlement := range items {
		products, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/entitlements/%s/products", escapedProjectID, url.PathEscape(entitlement.ID)), nil)
		if err != nil {
			return nil, fmt.Errorf("list products for entitlement %s: %w", entitlement.ID, err)
		}
		result = append(result, entitlementProducts{entitlement: entitlement, products: products})
	}
	return result, nil
}

func fetchOfferings(client *api.Client, escapedProjectID string) ([]offeringPackages, error) {
	items, err := api.PaginateAll[api.Offering](client, fmt.Sprintf("/projects/%s/offerings", escapedProjectID), nil)
	if err != nil {
		return nil, fmt.Errorf("list offerings: %w", err)
	}
	result := make([]offeringPackages, 0, len(items))
	for _, offering := range items {
		packages, err := fetchPackages(client, escapedProjectID, offering.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, offeringPackages{offering: offering, packages: packages})
	}
	return result, nil
}

func fetchPackages(client *api.Client, escapedProjectID, offeringID string) ([]packageProducts, error) {
	items, err := api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", escapedProjectID, url.PathEscape(offeringID)), nil)
	if err != nil {
		return nil, fmt.Errorf("list packages for offering %s: %w", offeringID, err)
	}
	result := make([]packageProducts, 0, len(items))
	for _, pkg := range items {
		products, err := api.PaginateAll[api.PackageProduct](client, fmt.Sprintf("/projects/%s/packages/%s/products", escapedProjectID, url.PathEscape(pkg.ID)), nil)
		if err != nil {
			return nil, fmt.Errorf("list products for package %s: %w", pkg.ID, err)
		}
		result = append(result, packageProducts{pkg: pkg, products: products})
	}
	return result, nil
}

func buildReport(projectID string, s *snapshot) *Report {
	report := &Report{
		Object:    "project_health_report",
		ProjectID: projectID,
		Status:    StatusPass,
	}

	activeProducts := map[string]api.Product{}
	allProducts := map[string]api.Product{}
	for _, product := range s.products {
		allProducts[product.ID] = product
		if isActive(product.State) {
			activeProducts[product.ID] = product
		}
	}

	entitlementProductIDs := map[string]bool{}
	entitlementProductLinks := 0
	activeEntitlements := 0
	for _, entitlement := range s.entitlements {
		if isActive(entitlement.entitlement.State) {
			activeEntitlements++
		}
		for _, product := range entitlement.products {
			entitlementProductLinks++
			entitlementProductIDs[product.ID] = true
		}
	}

	activeOfferings := 0
	currentOfferings := 0
	packageProductIDs := map[string]bool{}
	for _, offering := range s.offerings {
		if isActive(offering.offering.State) {
			activeOfferings++
		}
		if offering.offering.IsCurrent && isActive(offering.offering.State) {
			currentOfferings++
		}
		report.Counts.Packages += len(offering.packages)
		for _, pkg := range offering.packages {
			report.Counts.PackageProductLinks += len(pkg.products)
			for _, product := range pkg.products {
				packageProductIDs[product.ProductID] = true
			}
		}
	}

	report.Counts.Apps = len(s.apps)
	report.Counts.Products = len(s.products)
	report.Counts.ActiveProducts = len(activeProducts)
	report.Counts.Entitlements = len(s.entitlements)
	report.Counts.ActiveEntitlements = activeEntitlements
	report.Counts.EntitlementProductLinks = entitlementProductLinks
	report.Counts.Offerings = len(s.offerings)
	report.Counts.ActiveOfferings = activeOfferings
	report.Counts.CurrentOfferings = currentOfferings

	report.add(checkApps(s.apps))
	report.add(checkProducts(s.products, activeProducts))
	report.add(checkEntitlements(s.entitlements, activeProducts, entitlementProductIDs))
	report.add(checkOfferings(s.offerings, currentOfferings))
	report.add(checkPackages(s.offerings, allProducts, activeProducts, packageProductIDs))

	return report
}

func (r *Report) add(check Check) {
	r.Checks = append(r.Checks, check)
	applyStatus(&r.Status, check.Status)
}

func (r *LaunchReport) add(check Check) {
	r.Checks = append(r.Checks, check)
	applyStatus(&r.Status, check.Status)
	if check.Status == StatusFail {
		r.Ready = false
	}
}

func applyStatus(current *string, next string) {
	switch next {
	case StatusFail:
		*current = StatusFail
	case StatusWarn:
		if *current != StatusFail {
			*current = StatusWarn
		}
	}
}

func failingMessages(checks []Check) []string {
	return messagesByStatus(checks, StatusFail)
}

func warningMessages(checks []Check) []string {
	return messagesByStatus(checks, StatusWarn)
}

func messagesByStatus(checks []Check, status string) []string {
	var messages []string
	for _, check := range checks {
		if check.Status != status {
			continue
		}
		messages = append(messages, fmt.Sprintf("%s: %s", check.Area, check.Message))
		messages = append(messages, check.Details...)
	}
	return messages
}

func checkProjectHealth(report *Report) Check {
	switch report.Status {
	case StatusFail:
		return Check{Status: StatusFail, Area: "project", Message: "Project doctor found blocking setup errors.", Details: failingMessages(report.Checks)}
	case StatusWarn:
		return Check{Status: StatusWarn, Area: "project", Message: "Project doctor found warnings to review.", Details: warningMessages(report.Checks)}
	default:
		return Check{Status: StatusPass, Area: "project", Message: "Project doctor checks passed."}
	}
}

func checkLaunchCatalog(counts Counts) Check {
	switch {
	case counts.Apps == 0:
		return Check{Status: StatusFail, Area: "catalog", Message: "At least one app is required before launch."}
	case counts.ActiveProducts == 0:
		return Check{Status: StatusFail, Area: "catalog", Message: "At least one active product is required before launch."}
	default:
		return Check{Status: StatusPass, Area: "catalog", Message: fmt.Sprintf("%d app(s) and %d active product(s) are configured.", counts.Apps, counts.ActiveProducts)}
	}
}

func checkLaunchEntitlements(counts Counts) Check {
	switch {
	case counts.ActiveEntitlements == 0:
		return Check{Status: StatusFail, Area: "access", Message: "At least one active entitlement is required before launch."}
	case counts.EntitlementProductLinks == 0:
		return Check{Status: StatusFail, Area: "access", Message: "At least one product must be attached to an entitlement."}
	default:
		return Check{Status: StatusPass, Area: "access", Message: fmt.Sprintf("%d active entitlement(s) have product access paths.", counts.ActiveEntitlements)}
	}
}

func checkLaunchOffering(counts Counts) Check {
	switch {
	case counts.CurrentOfferings == 0:
		return Check{Status: StatusFail, Area: "offering", Message: "Exactly one current active offering is required before launch."}
	case counts.CurrentOfferings > 1:
		return Check{Status: StatusWarn, Area: "offering", Message: fmt.Sprintf("%d current active offerings found; expected exactly one.", counts.CurrentOfferings)}
	default:
		return Check{Status: StatusPass, Area: "offering", Message: "One current active offering is configured."}
	}
}

func checkLaunchPackages(counts Counts) Check {
	switch {
	case counts.Packages == 0:
		return Check{Status: StatusFail, Area: "paywall", Message: "The current offering needs at least one package before launch."}
	case counts.PackageProductLinks == 0:
		return Check{Status: StatusFail, Area: "paywall", Message: "At least one package must have an attached product before launch."}
	default:
		return Check{Status: StatusPass, Area: "paywall", Message: fmt.Sprintf("%d package(s) and %d package-product link(s) are configured.", counts.Packages, counts.PackageProductLinks)}
	}
}

func checkApps(apps []api.App) Check {
	if len(apps) == 0 {
		return Check{Status: StatusFail, Area: "apps", Message: "No apps found in this project."}
	}
	return Check{Status: StatusPass, Area: "apps", Message: fmt.Sprintf("%d app(s) configured.", len(apps))}
}

func checkProducts(products []api.Product, activeProducts map[string]api.Product) Check {
	switch {
	case len(products) == 0:
		return Check{Status: StatusFail, Area: "products", Message: "No products found in this project."}
	case len(activeProducts) == 0:
		return Check{Status: StatusFail, Area: "products", Message: "Products exist, but none are active."}
	default:
		return Check{Status: StatusPass, Area: "products", Message: fmt.Sprintf("%d active product(s) found.", len(activeProducts))}
	}
}

func checkEntitlements(entitlements []entitlementProducts, activeProducts map[string]api.Product, entitlementProductIDs map[string]bool) Check {
	if len(entitlements) == 0 {
		return Check{Status: StatusFail, Area: "entitlements", Message: "No entitlements found in this project."}
	}

	var details []string
	activeEntitlements := 0
	for _, entitlement := range entitlements {
		if !isActive(entitlement.entitlement.State) {
			continue
		}
		activeEntitlements++
		activeLinks := 0
		for _, product := range entitlement.products {
			if _, ok := activeProducts[product.ID]; ok {
				activeLinks++
			}
		}
		if activeLinks == 0 {
			details = append(details, fmt.Sprintf("%s has no active products attached", label(entitlement.entitlement.LookupKey, entitlement.entitlement.ID)))
		}
	}
	for productID, product := range activeProducts {
		if !entitlementProductIDs[productID] {
			details = append(details, fmt.Sprintf("%s is not attached to any entitlement", productLabel(product)))
		}
	}
	if activeEntitlements == 0 {
		return Check{Status: StatusFail, Area: "entitlements", Message: "Entitlements exist, but none are active."}
	}
	if len(details) > 0 {
		return Check{Status: StatusWarn, Area: "entitlements", Message: "Some entitlement/product links need attention.", Details: details}
	}
	return Check{Status: StatusPass, Area: "entitlements", Message: fmt.Sprintf("%d active entitlement(s) have product links.", activeEntitlements)}
}

func checkOfferings(offerings []offeringPackages, currentOfferings int) Check {
	if len(offerings) == 0 {
		return Check{Status: StatusFail, Area: "offerings", Message: "No offerings found in this project."}
	}
	if currentOfferings == 0 {
		return Check{Status: StatusFail, Area: "offerings", Message: "No current active offering is configured."}
	}
	if currentOfferings > 1 {
		return Check{Status: StatusWarn, Area: "offerings", Message: fmt.Sprintf("%d current active offerings found; expected exactly one.", currentOfferings)}
	}
	for _, offering := range offerings {
		if offering.offering.IsCurrent && isActive(offering.offering.State) {
			return Check{Status: StatusPass, Area: "offerings", Message: fmt.Sprintf("Current offering is %s.", label(offering.offering.LookupKey, offering.offering.ID))}
		}
	}
	return Check{Status: StatusFail, Area: "offerings", Message: "No current active offering is configured."}
}

func checkPackages(offerings []offeringPackages, allProducts, activeProducts map[string]api.Product, packageProductIDs map[string]bool) Check {
	var details []string
	currentPackageCount := 0
	currentProductLinks := 0
	for _, offering := range offerings {
		if !offering.offering.IsCurrent || !isActive(offering.offering.State) {
			continue
		}
		currentPackageCount += len(offering.packages)
		if len(offering.packages) == 0 {
			details = append(details, fmt.Sprintf("%s has no packages", label(offering.offering.LookupKey, offering.offering.ID)))
		}
		for _, pkg := range offering.packages {
			currentProductLinks += len(pkg.products)
			if len(pkg.products) == 0 {
				details = append(details, fmt.Sprintf("%s has no products attached", label(pkg.pkg.LookupKey, pkg.pkg.ID)))
				continue
			}
			for _, link := range pkg.products {
				product, ok := allProducts[link.ProductID]
				if !ok {
					details = append(details, fmt.Sprintf("%s references missing product %s", label(pkg.pkg.LookupKey, pkg.pkg.ID), link.ProductID))
					continue
				}
				if !isActive(product.State) {
					details = append(details, fmt.Sprintf("%s references inactive product %s", label(pkg.pkg.LookupKey, pkg.pkg.ID), productLabel(product)))
				}
			}
		}
	}
	for productID, product := range activeProducts {
		if !packageProductIDs[productID] {
			details = append(details, fmt.Sprintf("%s is not attached to any package", productLabel(product)))
		}
	}
	if currentPackageCount == 0 {
		return Check{Status: StatusFail, Area: "packages", Message: "The current offering has no packages.", Details: details}
	}
	if currentProductLinks == 0 {
		return Check{Status: StatusFail, Area: "packages", Message: "The current offering has packages, but no package products.", Details: details}
	}
	if len(details) > 0 {
		return Check{Status: StatusWarn, Area: "packages", Message: "Some offering package links need attention.", Details: details}
	}
	return Check{Status: StatusPass, Area: "packages", Message: fmt.Sprintf("Current offering has %d package(s) with product links.", currentPackageCount)}
}

func isActive(state string) bool {
	return state == "" || strings.EqualFold(state, "active")
}

func label(name, id string) string {
	if name == "" {
		return id
	}
	return fmt.Sprintf("%s (%s)", name, id)
}

func productLabel(product api.Product) string {
	return label(product.StoreIdentifier, product.ID)
}
