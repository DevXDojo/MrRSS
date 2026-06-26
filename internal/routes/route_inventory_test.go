package routes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestFlutterAPIRouteInventoryMatchesRegisteredRoutes(t *testing.T) {
	root := repositoryRoot(t)

	registered := registeredAPIRoutes(t, filepath.Join(root, "internal", "routes"))
	inventory := routeInventory(t, filepath.Join(root, "docs", "flutter", "API_ROUTE_INVENTORY.json"))

	if diff := compareStringSets(registered, inventory); diff != "" {
		t.Fatalf("Flutter API route inventory is out of sync with internal/routes:\n%s\nUpdate docs/flutter/API_ROUTE_INVENTORY.json and MIGRATION_LOG.md in the same migration slice.", diff)
	}
}

func registeredAPIRoutes(t *testing.T, routesDir string) []string {
	t.Helper()

	matches := map[string]bool{}
	re := regexp.MustCompile(`mux\.HandleFunc\("(/api/[^"]*)"`)

	entries, err := os.ReadDir(routesDir)
	if err != nil {
		t.Fatalf("read routes dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		body, err := os.ReadFile(filepath.Join(routesDir, entry.Name()))
		if err != nil {
			t.Fatalf("read route file %s: %v", entry.Name(), err)
		}

		for _, match := range re.FindAllStringSubmatch(string(body), -1) {
			matches[match[1]] = true
		}
	}

	routes := make([]string, 0, len(matches))
	for route := range matches {
		routes = append(routes, route)
	}
	sort.Strings(routes)

	return routes
}

func routeInventory(t *testing.T, path string) []string {
	t.Helper()

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read route inventory: %v", err)
	}

	var routes []string
	if err := json.Unmarshal(body, &routes); err != nil {
		t.Fatalf("parse route inventory: %v", err)
	}

	if !sort.StringsAreSorted(routes) {
		t.Fatalf("route inventory must be sorted alphabetically")
	}

	seen := map[string]bool{}
	for _, route := range routes {
		if !strings.HasPrefix(route, "/api/") {
			t.Fatalf("route inventory contains non-API route %q", route)
		}
		if seen[route] {
			t.Fatalf("route inventory contains duplicate route %q", route)
		}
		seen[route] = true
	}

	return routes
}

func compareStringSets(want, got []string) string {
	var missing []string
	var extra []string

	gotSet := map[string]bool{}
	for _, item := range got {
		gotSet[item] = true
	}
	for _, item := range want {
		if !gotSet[item] {
			missing = append(missing, item)
		}
	}

	wantSet := map[string]bool{}
	for _, item := range want {
		wantSet[item] = true
	}
	for _, item := range got {
		if !wantSet[item] {
			extra = append(extra, item)
		}
	}

	if len(missing) == 0 && len(extra) == 0 {
		return ""
	}

	var b strings.Builder
	if len(missing) > 0 {
		b.WriteString("missing from inventory:\n")
		for _, item := range missing {
			b.WriteString("  - ")
			b.WriteString(item)
			b.WriteByte('\n')
		}
	}
	if len(extra) > 0 {
		b.WriteString("extra in inventory:\n")
		for _, item := range extra {
			b.WriteString("  - ")
			b.WriteString(item)
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root from %s", dir)
		}
		dir = parent
	}
}
