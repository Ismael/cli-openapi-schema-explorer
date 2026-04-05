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

var depth int
var resolve bool
var pathFilter string

var pathsCmd = &cobra.Command{
	Use:   "paths [path] [method...]",
	Short: "Explore API paths progressively",
	Long: `List path segments, methods for a path, or full operation details.

Examples:
  openapi-explorer --spec api.yaml paths              # list top-level paths
  openapi-explorer --spec api.yaml paths --depth 2    # list paths two levels deep
  openapi-explorer --spec api.yaml paths /users        # list methods for /users
  openapi-explorer --spec api.yaml paths /users get    # full GET operation detail
  openapi-explorer --spec api.yaml paths /users get post  # multiple operations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		doc, err := loader.Load(specPath)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}

		allPaths := make([]string, 0)
		for p := range doc.Paths.Map() {
			allPaths = append(allPaths, p)
		}
		sort.Strings(allPaths)

		switch len(args) {
		case 0:
			return listPathSegments(cmd, doc, allPaths, depth)
		case 1:
			return showPath(cmd, doc, allPaths, args[0])
		default:
			return showOperation(cmd, doc, args[0], args[1:])
		}
	},
}

func init() {
	pathsCmd.Flags().IntVar(&depth, "depth", 0, "Path listing depth (0=auto)")
	pathsCmd.Flags().BoolVar(&resolve, "resolve", false, "Fully resolve all $ref inline")
	pathsCmd.Flags().StringVar(&pathFilter, "filter", "", "Filter paths by substring or tag name (case-insensitive)")
	rootCmd.AddCommand(pathsCmd)
}

const maxAutoLines = 50

func listPathSegments(cmd *cobra.Command, doc *loader.Doc, allPaths []string, depth int) error {
	// If filter is active, show all matching paths at full depth
	if cmd.Flags().Changed("filter") {
		filterVal, _ := cmd.Flags().GetString("filter")
		if filterVal != "" {
			filtered := filterPaths(doc, allPaths, filterVal)
			items := buildFullPathListing(doc, filtered)
			hint := fmt.Sprintf("Paths matching %q. Use: openapi-explorer paths <path> <method> for details", filterVal)
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(items, hint))
			return nil
		}
	}

	depthExplicit := cmd.Flags().Changed("depth")

	if depthExplicit && depth > 0 {
		items := buildPathListing(doc, allPaths, depth)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(items, "Use: openapi-explorer paths <path> <method> for details"))
		return nil
	}

	// Auto-depth: grow until exceeds maxAutoLines
	maxDepth := maxSegmentDepth(allPaths)
	bestDepth := 1
	bestItems := buildPathListing(doc, allPaths, 1)
	for d := 2; d <= maxDepth; d++ {
		items := buildPathListing(doc, allPaths, d)
		if len(items) > maxAutoLines {
			break
		}
		bestDepth = d
		bestItems = items
	}

	truncated := bestDepth < maxDepth
	if truncated {
		hint := fmt.Sprintf("Showing paths to depth %d. Use --depth N for more, or paths <path> to explore", bestDepth)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(bestItems, hint))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatList(bestItems, "Use: openapi-explorer paths <path> <method> for details"))
	}
	return nil
}

func buildPathListing(doc *loader.Doc, allPaths []string, depth int) []string {
	type segInfo struct {
		methods map[string]bool
	}
	segments := make(map[string]*segInfo)

	for _, p := range allPaths {
		parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
		d := depth
		if d > len(parts) {
			d = len(parts)
		}
		seg := "/" + strings.Join(parts[:d], "/")

		if _, exists := segments[seg]; !exists {
			segments[seg] = &segInfo{methods: make(map[string]bool)}
		}

		// Only add methods if this is an exact concrete path match
		if seg == p {
			pathItem := doc.Paths.Find(p)
			if pathItem != nil {
				for method := range pathItem.Operations() {
					segments[seg].methods[strings.ToUpper(method)] = true
				}
			}
		}
	}

	keys := make([]string, 0, len(segments))
	for k := range segments {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	items := make([]string, 0, len(keys))
	for _, seg := range keys {
		info := segments[seg]
		if len(info.methods) > 0 {
			methods := make([]string, 0, len(info.methods))
			for m := range info.methods {
				methods = append(methods, m)
			}
			sort.Strings(methods)
			items = append(items, strings.Join(methods, " ")+"  "+seg)
		} else {
			items = append(items, seg)
		}
	}
	return items
}

func maxSegmentDepth(allPaths []string) int {
	max := 0
	for _, p := range allPaths {
		parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
		if len(parts) > max {
			max = len(parts)
		}
	}
	return max
}

func showPath(cmd *cobra.Command, doc *loader.Doc, allPaths []string, path string) error {
	var methods []string
	var subPaths []string

	// Check if exact path exists
	pathItem := doc.Paths.Find(path)
	if pathItem != nil {
		for method, op := range pathItem.Operations() {
			summary := ""
			if op.Summary != "" {
				summary = ": " + op.Summary
			}
			methods = append(methods, strings.ToUpper(method)+summary)
		}
		sort.Strings(methods)
	}

	// Find sub-paths
	prefix := path
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for _, p := range allPaths {
		if strings.HasPrefix(p, prefix) && p != path {
			subPaths = append(subPaths, p)
		}
	}
	sort.Strings(subPaths)

	if len(methods) == 0 && len(subPaths) == 0 {
		msg := fmt.Sprintf("path %q not found in spec", path)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
		return fmt.Errorf("")
	}

	var lines []string
	if len(methods) > 0 {
		lines = append(lines, fmt.Sprintf("# Methods for %s", path))
		lines = append(lines, methods...)
	}
	if len(subPaths) > 0 {
		if len(methods) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "# Sub-paths")
		lines = append(lines, subPaths...)
	}

	fmt.Fprintln(cmd.OutOrStdout(), strings.Join(lines, "\n"))
	return nil
}

func showOperation(cmd *cobra.Command, doc *loader.Doc, path string, methods []string) error {
	pathItem := doc.Paths.Find(path)
	if pathItem == nil {
		msg := fmt.Sprintf("path %q not found in spec", path)
		fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
		return fmt.Errorf("")
	}

	// Build components map for resolver
	componentsMap := docComponentsToMap(doc)
	res := resolver.New(componentsMap)

	lower := make([]string, len(methods))
	for i, m := range methods {
		lower[i] = strings.ToLower(m)
	}
	var results []any

	for _, method := range lower {
		op := pathItem.GetOperation(strings.ToUpper(method))
		if op == nil {
			msg := fmt.Sprintf("method %q not found on path %q", strings.ToUpper(method), path)
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(msg))
			return fmt.Errorf("")
		}

		// Convert operation to generic map for ref resolution
		opJSON, err := json.Marshal(op)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), output.FormatError(err.Error()))
			return fmt.Errorf("")
		}
		var opMap map[string]any
		json.Unmarshal(opJSON, &opMap)

		// Also include path-level parameters if not overridden
		if pathItem.Parameters != nil && opMap["parameters"] == nil {
			paramsJSON, _ := json.Marshal(pathItem.Parameters)
			var params []any
			json.Unmarshal(paramsJSON, &params)
			opMap["parameters"] = params
		}

		var resolved any
		if cmd.Flags().Changed("resolve") {
			resolved = res.ResolveFull(opMap)
		} else {
			resolved = res.Resolve(opMap)
		}
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

func filterPaths(doc *loader.Doc, allPaths []string, filter string) []string {
	filterLower := strings.ToLower(filter)
	var result []string
	for _, p := range allPaths {
		// Match path substring
		if strings.Contains(strings.ToLower(p), filterLower) {
			result = append(result, p)
			continue
		}
		// Match operation tags
		pathItem := doc.Paths.Find(p)
		if pathItem != nil {
			matched := false
			for _, op := range pathItem.Operations() {
				for _, tag := range op.Tags {
					if strings.EqualFold(tag, filter) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if matched {
				result = append(result, p)
			}
		}
	}
	return result
}

func buildFullPathListing(doc *loader.Doc, paths []string) []string {
	items := make([]string, 0, len(paths))
	for _, p := range paths {
		pathItem := doc.Paths.Find(p)
		if pathItem != nil {
			methods := make([]string, 0)
			for method := range pathItem.Operations() {
				methods = append(methods, strings.ToUpper(method))
			}
			sort.Strings(methods)
			if len(methods) > 0 {
				items = append(items, strings.Join(methods, " ")+"  "+p)
			} else {
				items = append(items, p)
			}
		} else {
			items = append(items, p)
		}
	}
	return items
}

// docComponentsToMap converts the doc's components to a generic map for the resolver.
func docComponentsToMap(doc *loader.Doc) map[string]any {
	componentsJSON, err := json.Marshal(doc.Components)
	if err != nil {
		return map[string]any{}
	}
	var componentsMap map[string]any
	json.Unmarshal(componentsJSON, &componentsMap)
	return componentsMap
}
