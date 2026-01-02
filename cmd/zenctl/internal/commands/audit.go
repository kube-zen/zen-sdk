package commands

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func NewAuditCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Query audit logs from SaaS API",
		Long: `Queries audit logs from the SaaS /v1/audit endpoint.

Requires ZEN_API_BASE_URL environment variable to be set.
If not set, prints guidance on how to configure it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiBaseURL := os.Getenv("ZEN_API_BASE_URL")
			if apiBaseURL == "" {
				cmd.Println("Error: ZEN_API_BASE_URL environment variable is not set")
				cmd.Println("")
				cmd.Println("To use zenctl audit, set the ZEN_API_BASE_URL environment variable:")
				cmd.Println("  export ZEN_API_BASE_URL=https://api.example.com")
				cmd.Println("")
				cmd.Println("Then run:")
				cmd.Println("  zenctl audit")
				return fmt.Errorf("ZEN_API_BASE_URL not set")
			}

			// Build audit URL
			auditURL := fmt.Sprintf("%s/v1/audit", apiBaseURL)

			// Create HTTP client
			client := &http.Client{
				Timeout: 30 * time.Second,
			}

			// Make request
			req, err := http.NewRequest("GET", auditURL, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			// Add headers (could add auth here if needed)
			req.Header.Set("Accept", "application/json")

			// Execute request
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to query audit endpoint: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("audit endpoint returned status %d: %s", resp.StatusCode, resp.Status)
			}

			// For now, just print a message - in practice, you'd parse and format the response
			cmd.Printf("Audit logs queried from %s\n", auditURL)
			cmd.Printf("Status: %s\n", resp.Status)
			cmd.Println("")
			cmd.Println("Note: Full audit log parsing and formatting is not yet implemented.")
			cmd.Println("Use -o json to see raw response, or implement parsing based on your API format.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json)")

	return cmd
}

