package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	specPath string
)

var rootCmd = &cobra.Command{
	Use:   "openapi-explorer",
	Short: "CLI tool for exploring OpenAPI specifications progressively",
	Long: `A token-efficient CLI for AI agents to progressively discover and explore
OpenAPI (v3.0) and Swagger (v2.0) specifications.

Start with 'paths' or 'components' to discover the API structure.
Use comma-separated names for batch requests.
Output is YAML for readability.
All $ref references are resolved inline (use --resolve for full resolution).`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if specPath != "" {
			return nil
		}
		// Try reading config file
		data, err := os.ReadFile(".openapi-explorer")
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Error: no spec provided. Use --spec or create .openapi-explorer with a spec field")
			return fmt.Errorf("")
		}
		var config struct {
			Spec string `yaml:"spec"`
		}
		if err := yaml.Unmarshal(data, &config); err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), fmt.Sprintf("Error: invalid .openapi-explorer file: %v", err))
			return fmt.Errorf("")
		}
		if config.Spec == "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Error: .openapi-explorer file missing 'spec' field")
			return fmt.Errorf("")
		}
		specPath = config.Spec
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&specPath, "spec", "", "OpenAPI spec file path or URL")
	// NOTE: No longer marked as required — handled by PersistentPreRunE
}
