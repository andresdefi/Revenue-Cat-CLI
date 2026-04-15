package cmd_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestGoldenProductCreateRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "create",
		"--store-id", "com.example.premium.yearly",
		"--app-id", "app_cmdtest",
		"--type", "subscription",
		"--display-name", "Premium Yearly",
	})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/products", map[string]any{
		"store_identifier": "com.example.premium.yearly",
		"app_id":           "app_cmdtest",
		"type":             "subscription",
		"display_name":     "Premium Yearly",
	})
}

func TestGoldenProductUpdateRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "update", "prod_cmdtest", "--display-name", "Premium Monthly"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/products/prod_cmdtest", map[string]any{
		"display_name": "Premium Monthly",
	})
}

func TestGoldenEntitlementAttachRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "attach", "--entitlement-id", "entl_cmdtest", "--product-id", "prod_monthly,prod_yearly"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products", map[string]any{
		"product_ids": []any{"prod_monthly", "prod_yearly"},
	})
}

func TestGoldenPackageAttachRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "attach", "--package-id", "pkge_cmdtest", "--product-id", "prod_monthly", "--eligibility", "google_sdk_ge_6"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products", map[string]any{
		"products": []any{
			map[string]any{
				"product_id":           "prod_monthly",
				"eligibility_criteria": "google_sdk_ge_6",
			},
		},
	})
}
