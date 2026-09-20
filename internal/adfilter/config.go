// Package adfilter provides conservative, local, reversible HTML ad filtering.
package adfilter

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/andybalholm/cascadia"
)

type Options struct {
	Enabled              bool   `json:"enabled"`
	Allowlist            string `json:"allowlist"`
	Rules                string `json:"rules"`
	AIEnabled            bool   `json:"ai_enabled"`
	AIMode               string `json:"ai_mode"`
	IncludeSelfPromotion bool   `json:"include_self_promotion"`
}

type Config struct {
	Options
	Revision int64 `json:"revision"`
}

func DefaultConfig() Config { return Config{Options: Options{Enabled: true, AIMode: "review"}} }

type Rule struct {
	Domain    string
	Selector  cascadia.Selector
	Exception bool
	Line      int
}

var domainName = regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

func validDomain(domain string) bool {
	return len(domain) <= 253 && domainName.MatchString(domain)
}

// Custom cosmetic rules are deliberately domain-scoped. This supports ordinary
// CSS selectors, not remote lists, JavaScript snippets or arbitrary regexes.
func ParseRules(raw string) ([]Rule, error) {
	if len(raw) > 32*1024 {
		return nil, fmt.Errorf("rules_too_large")
	}
	rules := []Rule{}
	for index, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "!") {
			continue
		}
		separator := "##"
		exception := strings.Contains(line, "#@#")
		if exception {
			separator = "#@#"
		}
		parts := strings.SplitN(line, separator, 2)
		if len(parts) != 2 || !validDomain(strings.ToLower(parts[0])) || len(parts[1]) > 500 || strings.TrimSpace(parts[1]) == "" {
			return nil, fmt.Errorf("invalid_rule:%d", index+1)
		}
		// Keep expensive/emulated pseudo-selectors outside the supported subset.
		if strings.ContainsAny(parts[1], "{}\\") || strings.Contains(parts[1], ":has") || strings.Contains(parts[1], ":contains") || strings.Contains(parts[1], ":matches") {
			return nil, fmt.Errorf("invalid_rule:%d", index+1)
		}
		selector, err := cascadia.Compile(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid_rule:%d", index+1)
		}
		rules = append(rules, Rule{strings.ToLower(parts[0]), selector, exception, index + 1})
		if len(rules) > 100 {
			return nil, fmt.Errorf("too_many_rules")
		}
	}
	return rules, nil
}

func Validate(options Options) error {
	if options.AIMode != "" && options.AIMode != "review" && options.AIMode != "automatic" {
		return fmt.Errorf("invalid_ai_options")
	}
	if len(options.Allowlist) > 8192 {
		return fmt.Errorf("allowlist_too_large")
	}
	for _, line := range strings.Split(options.Allowlist, "\n") {
		domain := strings.ToLower(strings.TrimSpace(line))
		if domain != "" && !validDomain(domain) {
			return fmt.Errorf("invalid_domain")
		}
	}
	_, err := ParseRules(options.Rules)
	return err
}

func domainMatches(host, domain string) bool {
	return host == domain || strings.HasSuffix(host, "."+domain)
}

func Host(address string) string {
	parsed, err := url.Parse(address)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return ""
	}
	return strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
}

func (options Options) Applies(address string) bool {
	if !options.Enabled {
		return false
	}
	host := Host(address)
	if host == "" {
		return false
	}
	for _, line := range strings.Split(options.Allowlist, "\n") {
		domain := strings.ToLower(strings.TrimSpace(line))
		if domain != "" && domainMatches(host, domain) {
			return false
		}
	}
	return true
}

// Only dedicated ad-serving hosts; shared CDNs, editorial pages, analytics
// providers and substring matches are intentionally not included.
var adHosts = []string{"doubleclick.net", "googlesyndication.com", "googleadservices.com", "adnxs.com", "adsrvr.org", "gum.criteo.com", "bidder.criteo.com", "criteo.net", "ads.pubmatic.com", "adservice.google.com"}

func AdResource(address, base string) bool {
	root, err := url.Parse(base)
	if err != nil {
		return false
	}
	ref, err := url.Parse(strings.TrimSpace(address))
	if err != nil {
		return false
	}
	host := Host(root.ResolveReference(ref).String())
	for _, domain := range adHosts {
		if domainMatches(host, domain) {
			return true
		}
	}
	return false
}
