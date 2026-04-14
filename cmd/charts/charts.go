package charts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewChartsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "charts",
		Aliases: []string{"chart", "metrics"},
		Short:   "View charts and metrics",
	}
	root.AddCommand(newOverviewCmd(projectID, outputFormat))
	root.AddCommand(newShowCmd(projectID, outputFormat))
	root.AddCommand(newOptionsCmd(projectID, outputFormat))
	return root
}

func newOverviewCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Show metrics overview for a project",
		RunE: func(c *cobra.Command, args []string) error {
			run := func(_ context.Context) error {
				pid, err := cmdutil.ResolveProject(projectID)
				if err != nil {
					return err
				}
				client, err := api.NewClient()
				if err != nil {
					return err
				}
				data, err := client.Get(fmt.Sprintf("/projects/%s/metrics/overview", url.PathEscape(pid)), nil)
				if err != nil {
					return err
				}
				var metrics api.OverviewMetrics
				if err := json.Unmarshal(data, &metrics); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, metrics, func(t table.Writer) {
					t.AppendHeader(table.Row{"Metric", "Value", "Period", "Description"})
					for _, m := range metrics.Metrics {
						t.AppendRow(table.Row{m.Name, fmt.Sprintf("%.2f", m.Value), m.Period, m.Description})
					}
				})
				return nil
			}
			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func newShowCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "show <chart-name>", Short: "Show a specific chart's data", Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			data, err := client.Get(fmt.Sprintf("/projects/%s/charts/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var chart api.ChartData
			if err := json.Unmarshal(data, &chart); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, chart, func(t table.Writer) {
				t.SetTitle(chart.DisplayName)
				t.AppendHeader(table.Row{"Date", "Value"})
				for _, v := range chart.Values {
					t.AppendRow(table.Row{v.Date, fmt.Sprintf("%.2f", v.Value)})
				}
			})
			return nil
		},
	}
}

func newOptionsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "options <chart-name>", Short: "Get available filter/segment options for a chart", Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			data, err := client.Get(fmt.Sprintf("/projects/%s/charts/%s/options", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			if format == output.FormatJSON {
				var raw json.RawMessage
				if err := json.Unmarshal(data, &raw); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				output.Print(format, raw, nil)
			} else {
				var opts api.ChartOptions
				if err := json.Unmarshal(data, &opts); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				output.Print(format, opts, func(t table.Writer) {
					t.AppendHeader(table.Row{"Option", "Values"})
					for _, o := range opts.Options {
						t.AppendRow(table.Row{o.Name, fmt.Sprintf("%v", o.Values)})
					}
				})
			}
			return nil
		},
	}
}
