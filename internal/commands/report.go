package commands

import (
	"encoding/json"
	"fmt"

	"github.com/nicolasacchi/merchant-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	reportStartDate  string
	reportEndDate    string
	reportLimit      int
	reportCategoryID string
	reportCountry    string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Run Merchant Center reports (MCQL)",
}

var reportRunCmd = &cobra.Command{
	Use:   "run <mcql-query>",
	Short: "Run a raw MCQL report query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReport(args[0])
	},
}

var reportPerformanceCmd = &cobra.Command{
	Use:   "performance",
	Short: "Product performance report (clicks, impressions, CTR)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportStartDate == "" || reportEndDate == "" {
			return fmt.Errorf("--start-date and --end-date are required")
		}
		query := fmt.Sprintf(`
			SELECT
				product_performance_view.offer_id,
				product_performance_view.title,
				product_performance_view.clicks,
				product_performance_view.impressions,
				product_performance_view.click_through_rate
			FROM product_performance_view
			WHERE product_performance_view.date BETWEEN "%s" AND "%s"
			ORDER BY product_performance_view.clicks DESC
			LIMIT %d`, reportStartDate, reportEndDate, reportLimit)
		return runReport(query)
	},
}

var reportIssuesCmd = &cobra.Command{
	Use:   "issues",
	Short: "Product issues/disapprovals report",
	RunE: func(cmd *cobra.Command, args []string) error {
		query := fmt.Sprintf(`
			SELECT
				product_view.id,
				product_view.offer_id,
				product_view.title,
				product_view.aggregated_reporting_context_status,
				product_view.item_issues
			FROM product_view
			WHERE product_view.aggregated_reporting_context_status != "ELIGIBLE"
			LIMIT %d`, reportLimit)
		return runReport(query)
	},
}

var reportCompetitiveCmd = &cobra.Command{
	Use:   "competitive",
	Short: "Competitive visibility report",
	Long:  "Competitive visibility report. Requires --category-id (Google product category ID).",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reportStartDate == "" || reportEndDate == "" {
			return fmt.Errorf("--start-date and --end-date are required")
		}
		if reportCategoryID == "" {
			return fmt.Errorf("--category-id is required (Google product category ID)")
		}
		query := fmt.Sprintf(`
			SELECT
				competitive_visibility_competitor_view.report_category_id,
				competitive_visibility_competitor_view.report_country_code,
				competitive_visibility_competitor_view.traffic_source,
				competitive_visibility_competitor_view.domain,
				competitive_visibility_competitor_view.is_your_domain,
				competitive_visibility_competitor_view.relative_visibility,
				competitive_visibility_competitor_view.page_overlap_rate,
				competitive_visibility_competitor_view.higher_position_rate
			FROM competitive_visibility_competitor_view
			WHERE competitive_visibility_competitor_view.date BETWEEN "%s" AND "%s"
				AND competitive_visibility_competitor_view.report_category_id = %s
				AND competitive_visibility_competitor_view.report_country_code = "%s"
				AND competitive_visibility_competitor_view.traffic_source = "ORGANIC"
			ORDER BY competitive_visibility_competitor_view.relative_visibility DESC
			LIMIT %d`, reportStartDate, reportEndDate, reportCategoryID, reportCountry, reportLimit)
		return runReport(query)
	},
}

func runReport(query string) error {
	mid, err := requireMerchantID()
	if err != nil {
		return err
	}
	return runReportQuery(query, mid, 0, nil)
}

// runReportQuery executes an MCQL query; limit 0 means unlimited.
func runReportQuery(query, mid string, limit int, extra map[string]any) error {
	rows, err := newService().SearchReport(ctx(), "accounts/"+mid, query, limit)
	if err != nil {
		return WrapAPIError(fmt.Errorf("report search: %w", err))
	}
	raws := make([]json.RawMessage, 0, len(rows))
	for _, row := range rows {
		b, err := output.ProtoToRaw(row)
		if err != nil {
			return err
		}
		raws = append(raws, b)
	}
	out := map[string]any{"row_count": len(raws), "rows": raws}
	for k, v := range extra {
		out[k] = v
	}
	return output.PrintJSON(out)
}

func init() {
	reportPerformanceCmd.Flags().StringVar(&reportStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	reportPerformanceCmd.Flags().StringVar(&reportEndDate, "end-date", "", "End date (YYYY-MM-DD)")
	reportPerformanceCmd.Flags().IntVar(&reportLimit, "limit", 50, "Max rows")
	reportIssuesCmd.Flags().IntVar(&reportLimit, "limit", 100, "Max rows")
	reportCompetitiveCmd.Flags().StringVar(&reportStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	reportCompetitiveCmd.Flags().StringVar(&reportEndDate, "end-date", "", "End date (YYYY-MM-DD)")
	reportCompetitiveCmd.Flags().StringVar(&reportCategoryID, "category-id", "", "Google product category ID (required)")
	reportCompetitiveCmd.Flags().StringVar(&reportCountry, "country", "IT", "Country code (default: IT)")
	reportCompetitiveCmd.Flags().IntVar(&reportLimit, "limit", 50, "Max rows")

	reportCmd.AddCommand(reportRunCmd)
	reportCmd.AddCommand(reportPerformanceCmd)
	reportCmd.AddCommand(reportIssuesCmd)
	reportCmd.AddCommand(reportCompetitiveCmd)
}
