package cmd

import (
	"fmt"

	"github.com/Ismael/cli-openapi-schema-explorer/internal/loader"
	"github.com/Ismael/cli-openapi-schema-explorer/internal/output"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show spec metadata (title, version, description, servers)",
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := loader.Load(specPath)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}

		info := map[string]any{
			"title":   doc.Info.Title,
			"version": doc.Info.Version,
		}
		if doc.Info.Description != "" {
			info["description"] = doc.Info.Description
		}
		if doc.Servers != nil {
			servers := make([]map[string]any, 0, len(doc.Servers))
			for _, s := range doc.Servers {
				srv := map[string]any{"url": s.URL}
				if s.Description != "" {
					srv["description"] = s.Description
				}
				servers = append(servers, srv)
			}
			info["servers"] = servers
		}

		out, err := output.FormatYAML(info)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}
		fmt.Fprintln(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
