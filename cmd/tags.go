package cmd

import (
	"fmt"
	"sort"

	"github.com/Ismael/cli-openapi-schema-explorer/internal/loader"
	"github.com/Ismael/cli-openapi-schema-explorer/internal/output"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List API tags with descriptions",
	Long:  `List all tags defined in the spec with their descriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := loader.Load(specPath)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}

		var items []string
		if doc.Tags != nil {
			for _, tag := range doc.Tags {
				if tag.Description != "" {
					items = append(items, tag.Name+": "+tag.Description)
				} else {
					items = append(items, tag.Name)
				}
			}
		}
		sort.Strings(items)

		fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(items, "Tags in this spec. Use: paths --filter <tag> to see endpoints"))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
