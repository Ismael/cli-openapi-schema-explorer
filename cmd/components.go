package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Ismael/cli-openapi-schema-explorer/internal/loader"
	"github.com/Ismael/cli-openapi-schema-explorer/internal/output"
	"github.com/Ismael/cli-openapi-schema-explorer/internal/resolver"
	"github.com/spf13/cobra"
)

var filter string

var componentsCmd = &cobra.Command{
	Use:   "components [type] [name...]",
	Short: "Explore API components progressively",
	Long: `List component types, names within a type, or full component details.

Examples:
  openapi-explorer --spec api.yaml components                          # list types
  openapi-explorer --spec api.yaml components schemas                  # list schema names
  openapi-explorer --spec api.yaml components schemas --filter User    # filter names
  openapi-explorer --spec api.yaml components schemas User             # full schema detail
  openapi-explorer --spec api.yaml components schemas User Task        # multiple schemas`,
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := loader.Load(specPath)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}

		if doc.Components == nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatList([]string{}, "No components in this spec"))
			return nil
		}

		componentsMap := docComponentsToMap(doc)

		switch len(args) {
		case 0:
			return listComponentTypes(cmd, componentsMap)
		case 1:
			return listComponentNames(cmd, componentsMap, args[0], filter)
		default:
			return showComponentDetail(cmd, componentsMap, args[0], args[1:])
		}
	},
}

func init() {
	componentsCmd.Flags().StringVar(&filter, "filter", "", "Filter component names (case-insensitive substring)")
	rootCmd.AddCommand(componentsCmd)
}

func listComponentTypes(cmd *cobra.Command, componentsMap map[string]any) error {
	types := make([]string, 0)
	for k, v := range componentsMap {
		if m, ok := v.(map[string]any); ok && len(m) > 0 {
			types = append(types, k)
		}
	}
	sort.Strings(types)
	fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(types, "Use: openapi-explorer --spec <spec> components <type> to list names"))
	return nil
}

func listComponentNames(cmd *cobra.Command, componentsMap map[string]any, compType string, filter string) error {
	typeMap, ok := componentsMap[compType].(map[string]any)
	if !ok || len(typeMap) == 0 {
		available := make([]string, 0)
		for k, v := range componentsMap {
			if m, ok := v.(map[string]any); ok && len(m) > 0 {
				available = append(available, k)
			}
		}
		msg := fmt.Sprintf("unknown component type %q (available: %s)", compType, strings.Join(available, ", "))
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
		return fmt.Errorf("")
	}

	names := make([]string, 0, len(typeMap))
	filterLower := strings.ToLower(filter)
	for name := range typeMap {
		if filter == "" || strings.Contains(strings.ToLower(name), filterLower) {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	hint := fmt.Sprintf("Use: openapi-explorer --spec <spec> components %s <name> for details", compType)
	fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(names, hint))
	return nil
}

func showComponentDetail(cmd *cobra.Command, componentsMap map[string]any, compType string, names []string) error {
	typeMap, ok := componentsMap[compType].(map[string]any)
	if !ok || len(typeMap) == 0 {
		msg := fmt.Sprintf("unknown component type %q", compType)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
		return fmt.Errorf("")
	}

	res := resolver.New(componentsMap)
	var results []any

	for _, name := range names {
		schema, exists := typeMap[name]
		if !exists {
			msg := fmt.Sprintf("component %q not found in %s", name, compType)
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
			return fmt.Errorf("")
		}

		// Deep copy and resolve
		schemaJSON, _ := json.Marshal(schema)
		var schemaMap any
		json.Unmarshal(schemaJSON, &schemaMap)

		resolved := res.ResolveFull(schemaMap)
		results = append(results, resolved)
	}

	var out string
	var err error
	if len(results) == 1 {
		out, err = output.FormatYAML(results[0])
	} else {
		out, err = output.FormatYAML(results)
	}
	if err != nil {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
		return fmt.Errorf("")
	}
	fmt.Fprintln(cmd.OutOrStdout(), out)
	return nil
}
