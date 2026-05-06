package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const defaultSpecURL = "https://www.revenuecat.com/docs/redocusaurus/plugin-redoc-0.yaml"

type endpoint struct {
	Method string
	Path   string
}

type coverageRow struct {
	Endpoint    endpoint
	Implemented bool
	Tested      bool
	Documented  bool
}

func main() {
	specFlag := flag.String("spec", defaultSpecURL, "OpenAPI spec URL or local file path")
	outFlag := flag.String("out", "docs/API_COVERAGE.md", "coverage report output path")
	check := flag.Bool("check", false, "exit non-zero when spec endpoints are not implemented")
	flag.Parse()

	specData, err := readSpec(*specFlag)
	if err != nil {
		fail("read spec: %v", err)
	}
	specEndpoints, specFields := parseOpenAPIYAML(specData)
	if len(specEndpoints) == 0 {
		fail("no endpoints found in OpenAPI spec")
	}

	codeEndpoints, err := discoverCodeEndpoints(false)
	if err != nil {
		fail("discover code endpoints: %v", err)
	}
	testEndpoints, err := discoverCodeEndpoints(true)
	if err != nil {
		fail("discover test endpoints: %v", err)
	}
	docs, err := readDocs()
	if err != nil {
		fail("read docs: %v", err)
	}

	rows := buildRows(specEndpoints, codeEndpoints, testEndpoints, docs)
	fieldWarnings, err := fieldDriftWarnings(specFields)
	if err != nil {
		fail("check fields: %v", err)
	}
	if err := writeReport(*outFlag, *specFlag, rows, fieldWarnings); err != nil {
		fail("write report: %v", err)
	}

	missing := missingImplemented(rows)
	for _, warning := range fieldWarnings {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "Missing %d OpenAPI endpoint(s):\n", len(missing))
		for _, ep := range missing {
			fmt.Fprintf(os.Stderr, "  %s %s\n", ep.Method, ep.Path)
		}
		if *check {
			os.Exit(1)
		}
	}
}

func readSpec(spec string) ([]byte, error) {
	if strings.HasPrefix(spec, "http://") || strings.HasPrefix(spec, "https://") {
		resp, err := http.Get(spec) //nolint:gosec // user-controlled CI utility URL
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(spec) //nolint:gosec // user-specified spec path
}

func parseOpenAPIYAML(data []byte) ([]endpoint, map[string]map[string]struct{}) {
	var (
		inPaths       bool
		inSchemas     bool
		currentPath   string
		currentSchema string
		schemasAt     int
		propertiesAt  int
		endpoints     []endpoint
		fields        = map[string]map[string]struct{}{}
	)
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "paths:" {
			inPaths = true
			inSchemas = false
			currentPath = ""
			continue
		}
		if strings.HasPrefix(trimmed, "components:") || strings.HasPrefix(trimmed, "x-tagGroups:") {
			inPaths = false
			currentPath = ""
		}
		if trimmed == "schemas:" {
			inSchemas = true
			inPaths = false
			schemasAt = leadingSpaces(line)
			continue
		}
		if inPaths {
			indent := leadingSpaces(line)
			if indent == 2 && strings.HasPrefix(trimmed, "/") && strings.HasSuffix(trimmed, ":") {
				currentPath = cleanYAMLKey(trimmed)
				currentPath = strings.TrimPrefix(currentPath, "/v2")
				continue
			}
			if currentPath != "" && indent == 4 && strings.HasSuffix(trimmed, ":") {
				method := strings.TrimSuffix(trimmed, ":")
				switch method {
				case "get", "post", "delete", "put", "patch":
					endpoints = append(endpoints, endpoint{Method: strings.ToUpper(method), Path: currentPath})
				}
			}
		}
		if inSchemas {
			indent := leadingSpaces(line)
			if indent == schemasAt+2 && strings.HasSuffix(trimmed, ":") {
				currentSchema = cleanYAMLKey(trimmed)
				propertiesAt = 0
				fields[currentSchema] = map[string]struct{}{}
				continue
			}
			if currentSchema != "" && propertiesAt == 0 && trimmed == "properties:" {
				propertiesAt = indent
				continue
			}
			if currentSchema != "" && propertiesAt > 0 && indent == propertiesAt+2 && strings.HasSuffix(trimmed, ":") {
				fields[currentSchema][cleanYAMLKey(trimmed)] = struct{}{}
			}
			if indent <= 2 && trimmed != "" && !strings.HasSuffix(trimmed, ":") {
				currentSchema = ""
				propertiesAt = 0
			}
		}
	}
	sortEndpoints(endpoints)
	return endpoints, fields
}

func leadingSpaces(s string) int {
	return len(s) - len(strings.TrimLeft(s, " "))
}

func cleanYAMLKey(s string) string {
	s = strings.TrimSuffix(strings.TrimSpace(s), ":")
	s = strings.Trim(s, `"'`)
	return s
}

func discoverCodeEndpoints(includeTests bool) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	dirs := []string{"cmd", "internal"}
	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			isTest := strings.HasSuffix(path, "_test.go")
			if includeTests != isTest {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, ep := range endpointsFromGoFile(path, data) {
				out[endpointKey(ep)] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func endpointsFromGoFile(path string, data []byte) []endpoint {
	file, err := parser.ParseFile(token.NewFileSet(), path, data, 0)
	if err != nil {
		return endpointsFromText(data)
	}
	var endpoints []endpoint
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		endpoints = append(endpoints, endpointsFromBlock(fn.Body)...)
	}
	return endpoints
}

func endpointsFromBlock(body *ast.BlockStmt) []endpoint {
	var endpoints []endpoint
	values := map[string]string{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch typed := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range typed.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || i >= len(typed.Rhs) {
					continue
				}
				if value, ok := stringFromExpr(typed.Rhs[i], values); ok {
					values[ident.Name] = value
				}
			}
		case *ast.ValueSpec:
			for i, name := range typed.Names {
				if i >= len(typed.Values) {
					continue
				}
				if value, ok := stringFromExpr(typed.Values[i], values); ok {
					values[name.Name] = value
				}
			}
		case *ast.CallExpr:
			if ep, ok := endpointFromCall(typed, values); ok {
				endpoints = append(endpoints, ep)
			}
		}
		return true
	})
	return endpoints
}

func endpointFromCall(call *ast.CallExpr, values map[string]string) (endpoint, bool) {
	if ep, ok := endpointFromAssertRequested(call, values); ok {
		return ep, true
	}
	method := ""
	switch selectorName(call.Fun) {
	case "Get", "GetFullURL":
		method = "GET"
	case "Post":
		method = "POST"
	case "Delete":
		method = "DELETE"
	case "Paginate", "PaginateAll":
		method = "GET"
	}
	if method == "" {
		return endpoint{}, false
	}
	if len(call.Args) == 0 {
		return endpoint{}, false
	}
	pathArg := call.Args[0]
	if method == "GET" && isPaginateCall(call) {
		if len(call.Args) < 2 {
			return endpoint{}, false
		}
		pathArg = call.Args[1]
	}
	path, ok := stringFromExpr(pathArg, values)
	if !ok || !strings.HasPrefix(path, "/") {
		return endpoint{}, false
	}
	return endpoint{Method: method, Path: normalizePath(path)}, true
}

func endpointFromAssertRequested(call *ast.CallExpr, values map[string]string) (endpoint, bool) {
	if selectorName(call.Fun) != "AssertRequested" || len(call.Args) < 4 {
		return endpoint{}, false
	}
	method, ok := stringFromExpr(call.Args[2], values)
	if !ok {
		return endpoint{}, false
	}
	path, ok := stringFromExpr(call.Args[3], values)
	if !ok || !strings.HasPrefix(path, "/") {
		return endpoint{}, false
	}
	return endpoint{Method: strings.ToUpper(method), Path: normalizePath(path)}, true
}

func isPaginateCall(call *ast.CallExpr) bool {
	name := selectorName(call.Fun)
	return name == "Paginate" || name == "PaginateAll"
}

func selectorName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.IndexExpr:
		return selectorName(typed.X)
	case *ast.IndexListExpr:
		return selectorName(typed.X)
	}
	return ""
}

func stringFromExpr(expr ast.Expr, values map[string]string) (string, bool) {
	switch typed := expr.(type) {
	case *ast.BasicLit:
		if typed.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(typed.Value)
		return value, err == nil
	case *ast.Ident:
		value, ok := values[typed.Name]
		return value, ok
	case *ast.SelectorExpr:
		switch typed.Sel.Name {
		case "MethodDelete":
			return "DELETE", true
		case "MethodGet":
			return "GET", true
		case "MethodPost":
			return "POST", true
		case "MethodPut":
			return "PUT", true
		case "MethodPatch":
			return "PATCH", true
		}
	case *ast.CallExpr:
		if selector, ok := typed.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Sprintf" && len(typed.Args) > 0 {
			return stringFromExpr(typed.Args[0], values)
		}
	case *ast.BinaryExpr:
		left, okLeft := stringFromExpr(typed.X, values)
		right, okRight := stringFromExpr(typed.Y, values)
		if okLeft && okRight && typed.Op.String() == "+" {
			return left + right, true
		}
	}
	return "", false
}

var (
	paramRE        = regexp.MustCompile(`\{[^}/]+\}`)
	printfRE       = regexp.MustCompile(`%[sdv]`)
	testPathRE     = regexp.MustCompile(`"(/[A-Za-z0-9_.$%{}=:/?&+-]+)"`)
	quotedPathRE   = regexp.MustCompile("`(/[A-Za-z0-9_.$%{}=:/?&+-]+)`")
	paramSegmentRE = regexp.MustCompile(`^(app|cust|entl|invoice|ofrnge|paywall|pkge|prod|proj|purch|sub|txn|vc|wh)[0-9_][A-Za-z0-9_-]*$`)
)

func endpointsFromText(data []byte) []endpoint {
	var endpoints []endpoint
	text := string(data)
	for _, match := range append(testPathRE.FindAllStringSubmatch(text, -1), quotedPathRE.FindAllStringSubmatch(text, -1)...) {
		if len(match) < 2 {
			continue
		}
		endpoints = append(endpoints, endpoint{Method: "GET", Path: normalizePath(match[1])})
	}
	return endpoints
}

func normalizePath(path string) string {
	path = strings.TrimPrefix(path, "/v2")
	if i := strings.Index(path, "?"); i >= 0 {
		path = path[:i]
	}
	path = printfRE.ReplaceAllString(path, "{}")
	path = paramRE.ReplaceAllString(path, "{}")
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if shouldNormalizeSegment(part) || (i > 0 && parts[i-1] == "charts") {
			parts[i] = "{}"
		}
	}
	return strings.Join(parts, "/")
}

func shouldNormalizeSegment(segment string) bool {
	if segment == "" || segment == "{}" {
		return false
	}
	if segment == "user-123" || strings.Contains(segment, "_cmdtest") {
		return true
	}
	if paramSegmentRE.MatchString(segment) {
		return true
	}
	if strings.ToUpper(segment) == segment && strings.ContainsAny(segment, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return true
	}
	return false
}

func endpointKey(ep endpoint) string {
	return ep.Method + " " + normalizePath(ep.Path)
}

func buildRows(spec []endpoint, code, tests map[string]struct{}, docs string) []coverageRow {
	rows := make([]coverageRow, 0, len(spec))
	for _, ep := range spec {
		key := endpointKey(ep)
		_, implemented := code[key]
		_, tested := tests[key]
		rows = append(rows, coverageRow{
			Endpoint:    ep,
			Implemented: implemented,
			Tested:      tested,
			Documented:  documented(ep, docs),
		})
	}
	return rows
}

func documented(ep endpoint, docs string) bool {
	key := documentationKey(ep.Path)
	if key == "" {
		return false
	}
	normalizedDocs := strings.ReplaceAll(strings.ToLower(docs), "-", "_")
	for _, candidate := range []string{key, strings.ReplaceAll(key, "_", " ")} {
		if strings.Contains(normalizedDocs, " "+candidate+" ") || strings.Contains(normalizedDocs, "rc "+candidate) {
			return true
		}
	}
	return false
}

func documentationKey(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" || strings.HasPrefix(part, "{") {
			continue
		}
		if part == "projects" && i == len(parts)-1 {
			return "projects"
		}
		if part == "projects" {
			continue
		}
		if part == "integrations" && i+1 < len(parts) {
			return parts[i+1]
		}
		if part == "metrics" {
			return "charts"
		}
		if part == "virtual_currencies" {
			return "currencies"
		}
		return part
	}
	return ""
}

func readDocs() (string, error) {
	var b strings.Builder
	for _, file := range []string{"README.md", "docs/COMMANDS.md", "docs/API_NOTES.md", "docs/WORKFLOWS.md"} {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	return strings.ToLower(b.String()), nil
}

func fieldDriftWarnings(specFields map[string]map[string]struct{}) ([]string, error) {
	data, err := os.ReadFile("internal/api/types.go")
	if err != nil {
		return nil, err
	}
	localFields := map[string]map[string]struct{}{}
	typeRE := regexp.MustCompile(`^type ([A-Za-z0-9_]+) struct \{`)
	jsonRE := regexp.MustCompile("`json:\"([^\",]+)")
	current := ""
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if match := typeRE.FindStringSubmatch(line); len(match) == 2 {
			current = match[1]
			localFields[current] = map[string]struct{}{}
			continue
		}
		if current != "" && line == "}" {
			current = ""
			continue
		}
		if current != "" {
			if match := jsonRE.FindStringSubmatch(line); len(match) == 2 && match[1] != "-" {
				localFields[current][match[1]] = struct{}{}
			}
		}
	}
	var warnings []string
	for schema, spec := range specFields {
		local, ok := localFields[schema]
		if !ok || len(spec) == 0 {
			continue
		}
		var missing []string
		for field := range spec {
			if _, ok := local[field]; !ok {
				missing = append(missing, field)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			warnings = append(warnings, fmt.Sprintf("%s missing local JSON field(s): %s", schema, strings.Join(missing, ", ")))
		}
	}
	sort.Strings(warnings)
	return warnings, nil
}

func writeReport(path, specSource string, rows []coverageRow, fieldWarnings []string) error {
	var b strings.Builder
	total := len(rows)
	implemented := total - len(missingImplemented(rows))
	tested := 0
	documented := 0
	for _, row := range rows {
		if row.Tested {
			tested++
		}
		if row.Documented {
			documented++
		}
	}
	fmt.Fprintf(&b, "# RevenueCat API Coverage\n\n")
	fmt.Fprintf(&b, "Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "Spec source: `%s`\n\n", specSource)
	fmt.Fprintf(&b, "Summary: %d/%d implemented, %d/%d tested, %d/%d documented.\n\n", implemented, total, tested, total, documented, total)
	fmt.Fprintf(&b, "| Method | Endpoint | Implemented | Tested | Documented |\n")
	fmt.Fprintf(&b, "|--------|----------|-------------|--------|------------|\n")
	for _, row := range rows {
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s | %s |\n", row.Endpoint.Method, row.Endpoint.Path, mark(row.Implemented), mark(row.Tested), mark(row.Documented))
	}
	if len(fieldWarnings) > 0 {
		fmt.Fprintf(&b, "\n## Field Drift Warnings\n\n")
		for _, warning := range fieldWarnings {
			fmt.Fprintf(&b, "- %s\n", warning)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func mark(ok bool) string {
	if ok {
		return "yes"
	}
	return "no"
}

func missingImplemented(rows []coverageRow) []endpoint {
	var missing []endpoint
	for _, row := range rows {
		if !row.Implemented {
			missing = append(missing, row.Endpoint)
		}
	}
	sortEndpoints(missing)
	return missing
}

func sortEndpoints(endpoints []endpoint) {
	sort.Slice(endpoints, func(i, j int) bool {
		return endpointKey(endpoints[i]) < endpointKey(endpoints[j])
	})
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
