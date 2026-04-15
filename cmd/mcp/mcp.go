package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/auth"
	"github.com/andresdefi/rc/internal/cmdutil"
	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// NewMCPCmd returns the top-level `mcp` command group.
func NewMCPCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server for RevenueCat operations",
	}
	cmdutil.MarkExperimental(root)
	root.AddCommand(newServeCmd())
	return root
}

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start an MCP server over stdio",
		Long: `Start a Model Context Protocol server that exposes RevenueCat CLI
operations as MCP tools. The server communicates over stdin/stdout.

The API key is resolved in this order:
  1. RC_API_KEY environment variable
  2. Stored keychain / config token (same as normal CLI auth)`,
		Example: `  # Start the MCP server
  rc mcp serve

  # Start with an explicit API key
  RC_API_KEY=sk_test_xxx rc mcp serve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := resolveClient()
			if err != nil {
				return err
			}
			return runServer(client)
		},
	}
}

func resolveClient() (*api.Client, error) {
	if key := os.Getenv("RC_API_KEY"); key != "" {
		return api.NewClientWithToken(key), nil
	}
	token, err := auth.GetToken("")
	if err != nil {
		return nil, fmt.Errorf("no API key found: set RC_API_KEY or run `rc auth login`")
	}
	return api.NewClientWithToken(token), nil
}

// jsonText marshals v to indented JSON and wraps it in a CallToolResult.
func jsonText(v any) (*gomcp.CallToolResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil
}

// errResult returns a tool-level error (not a protocol error).
func errResult(err error) (*gomcp.CallToolResult, any, error) {
	r := &gomcp.CallToolResult{IsError: true}
	r.SetError(err)
	return r, nil, nil
}

func runServer(client *api.Client) error {
	server := gomcp.NewServer(&gomcp.Implementation{
		Name:    "rc-mcp",
		Version: "0.1.0",
	}, nil)

	registerProjectTools(server, client)
	registerProductTools(server, client)
	registerEntitlementTools(server, client)
	registerOfferingTools(server, client)
	registerCustomerTools(server, client)
	registerSubscriptionTools(server, client)
	registerMetricsTools(server, client)

	return server.Run(context.Background(), &gomcp.StdioTransport{})
}

// --- Projects ---

type ListProjectsInput struct{}

func registerProjectTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_projects",
		Description: "List all RevenueCat projects accessible with the current API key",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, _ ListProjectsInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get("/projects", nil)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.Project]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})
}

// --- Products ---

type ListProductsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
	AppID     string `json:"app_id,omitempty" jsonschema:"description=Optional app ID filter"`
}

type GetProductInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
	ProductID string `json:"product_id" jsonschema:"required,description=The product ID"`
}

type CreateProductInput struct {
	ProjectID       string `json:"project_id" jsonschema:"required,description=The project ID"`
	StoreIdentifier string `json:"store_identifier" jsonschema:"required,description=Store product identifier"`
	AppID           string `json:"app_id" jsonschema:"required,description=App ID"`
	Type            string `json:"type" jsonschema:"required,description=Product type: subscription or one_time"`
	DisplayName     string `json:"display_name,omitempty" jsonschema:"description=Optional display name"`
}

func registerProductTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_products",
		Description: "List products in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args ListProductsInput) (*gomcp.CallToolResult, any, error) {
		query := url.Values{}
		if args.AppID != "" {
			query.Set("app_id", args.AppID)
		}
		data, err := client.Get(fmt.Sprintf("/projects/%s/products", url.PathEscape(args.ProjectID)), query)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.Product]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_product",
		Description: "Get a product by ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args GetProductInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/products/%s", url.PathEscape(args.ProjectID), url.PathEscape(args.ProductID)), nil)
		if err != nil {
			return errResult(err)
		}
		var product api.Product
		if err := json.Unmarshal(data, &product); err != nil {
			return errResult(err)
		}
		r, err := jsonText(product)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "create_product",
		Description: "Create a new product in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args CreateProductInput) (*gomcp.CallToolResult, any, error) {
		body := map[string]any{
			"store_identifier": args.StoreIdentifier,
			"app_id":           args.AppID,
			"type":             args.Type,
		}
		if args.DisplayName != "" {
			body["display_name"] = args.DisplayName
		}
		data, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(args.ProjectID)), body)
		if err != nil {
			return errResult(err)
		}
		var product api.Product
		if err := json.Unmarshal(data, &product); err != nil {
			return errResult(err)
		}
		r, err := jsonText(product)
		return r, nil, err
	})
}

// --- Entitlements ---

type ListEntitlementsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
}

type GetEntitlementInput struct {
	ProjectID     string `json:"project_id" jsonschema:"required,description=The project ID"`
	EntitlementID string `json:"entitlement_id" jsonschema:"required,description=The entitlement ID"`
}

type CreateEntitlementInput struct {
	ProjectID   string `json:"project_id" jsonschema:"required,description=The project ID"`
	LookupKey   string `json:"lookup_key" jsonschema:"required,description=Lookup key identifier"`
	DisplayName string `json:"display_name" jsonschema:"required,description=Display name"`
}

func registerEntitlementTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_entitlements",
		Description: "List entitlements in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args ListEntitlementsInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(args.ProjectID)), nil)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.Entitlement]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_entitlement",
		Description: "Get an entitlement by ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args GetEntitlementInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/entitlements/%s", url.PathEscape(args.ProjectID), url.PathEscape(args.EntitlementID)), nil)
		if err != nil {
			return errResult(err)
		}
		var ent api.Entitlement
		if err := json.Unmarshal(data, &ent); err != nil {
			return errResult(err)
		}
		r, err := jsonText(ent)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "create_entitlement",
		Description: "Create a new entitlement in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args CreateEntitlementInput) (*gomcp.CallToolResult, any, error) {
		body := map[string]any{
			"lookup_key":   args.LookupKey,
			"display_name": args.DisplayName,
		}
		data, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(args.ProjectID)), body)
		if err != nil {
			return errResult(err)
		}
		var ent api.Entitlement
		if err := json.Unmarshal(data, &ent); err != nil {
			return errResult(err)
		}
		r, err := jsonText(ent)
		return r, nil, err
	})
}

// --- Offerings ---

type ListOfferingsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
}

type GetOfferingInput struct {
	ProjectID  string `json:"project_id" jsonschema:"required,description=The project ID"`
	OfferingID string `json:"offering_id" jsonschema:"required,description=The offering ID"`
}

func registerOfferingTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_offerings",
		Description: "List offerings in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args ListOfferingsInput) (*gomcp.CallToolResult, any, error) {
		query := url.Values{}
		query.Set("expand", "items.package")
		data, err := client.Get(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(args.ProjectID)), query)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.Offering]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_offering",
		Description: "Get an offering by ID with its packages",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args GetOfferingInput) (*gomcp.CallToolResult, any, error) {
		query := url.Values{}
		query.Set("expand", "package,package.product")
		data, err := client.Get(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(args.ProjectID), url.PathEscape(args.OfferingID)), query)
		if err != nil {
			return errResult(err)
		}
		var offering api.Offering
		if err := json.Unmarshal(data, &offering); err != nil {
			return errResult(err)
		}
		r, err := jsonText(offering)
		return r, nil, err
	})
}

// --- Customers ---

type LookupCustomerInput struct {
	ProjectID  string `json:"project_id" jsonschema:"required,description=The project ID"`
	CustomerID string `json:"customer_id" jsonschema:"required,description=The customer/app user ID"`
}

type ListCustomerEntitlementsInput struct {
	ProjectID  string `json:"project_id" jsonschema:"required,description=The project ID"`
	CustomerID string `json:"customer_id" jsonschema:"required,description=The customer/app user ID"`
}

type GrantEntitlementInput struct {
	ProjectID     string `json:"project_id" jsonschema:"required,description=The project ID"`
	CustomerID    string `json:"customer_id" jsonschema:"required,description=The customer/app user ID"`
	EntitlementID string `json:"entitlement_id" jsonschema:"required,description=The entitlement ID to grant"`
	ExpiresAt     int64  `json:"expires_at" jsonschema:"required,description=Expiration timestamp in ms since epoch"`
}

func registerCustomerTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "lookup_customer",
		Description: "Look up a customer by their app user ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args LookupCustomerInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s", url.PathEscape(args.ProjectID), url.PathEscape(args.CustomerID)), nil)
		if err != nil {
			return errResult(err)
		}
		var customer api.Customer
		if err := json.Unmarshal(data, &customer); err != nil {
			return errResult(err)
		}
		r, err := jsonText(customer)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_customer_entitlements",
		Description: "List active entitlements for a customer",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args ListCustomerEntitlementsInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s/active_entitlements", url.PathEscape(args.ProjectID), url.PathEscape(args.CustomerID)), nil)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.ActiveEntitlement]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "grant_entitlement",
		Description: "Grant an entitlement to a customer (creates a promotional subscription)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args GrantEntitlementInput) (*gomcp.CallToolResult, any, error) {
		_, err := client.Post(
			fmt.Sprintf("/projects/%s/customers/%s/actions/grant_entitlement", url.PathEscape(args.ProjectID), url.PathEscape(args.CustomerID)),
			map[string]any{"entitlement_id": args.EntitlementID, "expires_at": args.ExpiresAt},
		)
		if err != nil {
			return errResult(err)
		}
		r, err := jsonText(map[string]string{"status": "granted", "entitlement_id": args.EntitlementID, "customer_id": args.CustomerID})
		return r, nil, err
	})
}

// --- Subscriptions ---

type ListSubscriptionsInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
}

type GetSubscriptionInput struct {
	ProjectID      string `json:"project_id" jsonschema:"required,description=The project ID"`
	SubscriptionID string `json:"subscription_id" jsonschema:"required,description=The subscription ID"`
}

func registerSubscriptionTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "list_subscriptions",
		Description: "List subscriptions in a RevenueCat project",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args ListSubscriptionsInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions", url.PathEscape(args.ProjectID)), nil)
		if err != nil {
			return errResult(err)
		}
		var resp api.ListResponse[api.Subscription]
		if err := json.Unmarshal(data, &resp); err != nil {
			return errResult(err)
		}
		r, err := jsonText(resp)
		return r, nil, err
	})

	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "get_subscription",
		Description: "Get a subscription by ID",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args GetSubscriptionInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions/%s", url.PathEscape(args.ProjectID), url.PathEscape(args.SubscriptionID)), nil)
		if err != nil {
			return errResult(err)
		}
		var sub api.Subscription
		if err := json.Unmarshal(data, &sub); err != nil {
			return errResult(err)
		}
		r, err := jsonText(sub)
		return r, nil, err
	})
}

// --- Metrics ---

type MetricsOverviewInput struct {
	ProjectID string `json:"project_id" jsonschema:"required,description=The project ID"`
}

func registerMetricsTools(server *gomcp.Server, client *api.Client) {
	gomcp.AddTool(server, &gomcp.Tool{
		Name:        "metrics_overview",
		Description: "Get the metrics overview for a RevenueCat project (MRR, active subs, revenue, etc.)",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, args MetricsOverviewInput) (*gomcp.CallToolResult, any, error) {
		data, err := client.Get(fmt.Sprintf("/projects/%s/metrics/overview", url.PathEscape(args.ProjectID)), nil)
		if err != nil {
			return errResult(err)
		}
		var metrics api.OverviewMetrics
		if err := json.Unmarshal(data, &metrics); err != nil {
			return errResult(err)
		}
		r, err := jsonText(metrics)
		return r, nil, err
	})
}
