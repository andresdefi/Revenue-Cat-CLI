package completions

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cache"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/spf13/cobra"
)

// WithCompletion sets ValidArgsFunction on a command and returns it for chaining.
func WithCompletion(cmd *cobra.Command, fn func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) *cobra.Command {
	cmd.ValidArgsFunction = fn
	return cmd
}

// completeResourceIDs fetches resource IDs from the API for shell completion.
// Uses the cache to avoid hitting the API on every tab press.
func completeResourceIDs(projectFlag *string, pathFmt string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		pid, err := cmdutil.ResolveProject(projectFlag)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cacheKey := fmt.Sprintf("completion:%s:%s", pid, pathFmt)
		var items []map[string]any

		if cached := cache.Get(cacheKey); cached != nil {
			_ = json.Unmarshal(cached, &items)
		}

		if items == nil {
			client, err := api.NewClient()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			path := fmt.Sprintf(pathFmt, url.PathEscape(pid))
			data, err := client.Get(path, nil)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			var resp struct {
				Items []map[string]any `json:"items"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items = resp.Items
			if raw, err := json.Marshal(items); err == nil {
				cache.Set(cacheKey, raw)
			}
		}

		return formatCompletions(items), cobra.ShellCompDirectiveNoFileComp
	}
}

func formatCompletions(items []map[string]any) []string {
	var out []string
	for _, item := range items {
		id, _ := item["id"].(string)
		if id == "" {
			continue
		}
		desc := firstNonEmpty(item, "display_name", "name", "lookup_key", "store_identifier")
		if desc != "" {
			out = append(out, fmt.Sprintf("%s\t%s", id, desc))
		} else {
			out = append(out, id)
		}
	}
	return out
}

func firstNonEmpty(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

func ProductIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/products")
}

func EntitlementIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/entitlements")
}

func OfferingIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/offerings")
}

func PackageIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/packages")
}

func AppIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/apps")
}

func WebhookIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/integrations/webhooks")
}

func SubscriptionIDs(projectFlag *string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return completeResourceIDs(projectFlag, "/projects/%s/subscriptions")
}

func ProjectIDs() func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		cacheKey := "completion:projects"
		var items []map[string]any

		if cached := cache.Get(cacheKey); cached != nil {
			_ = json.Unmarshal(cached, &items)
		}

		if items == nil {
			client, err := api.NewClient()
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			data, err := client.Get("/projects", nil)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var resp struct {
				Items []map[string]any `json:"items"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			items = resp.Items
			if raw, err := json.Marshal(items); err == nil {
				cache.Set(cacheKey, raw)
			}
		}

		return formatCompletions(items), cobra.ShellCompDirectiveNoFileComp
	}
}
