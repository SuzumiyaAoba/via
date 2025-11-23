package matcher

import (
	"net/url"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/gabriel-vasile/mimetype"
)

func Match(rules []config.Rule, filename string) ([]*config.Rule, error) {
	var matches []*config.Rule
	
	for i := range rules {
		rule := &rules[i]
		matched, err := matchRule(rule, filename)
		if err != nil {
			return nil, err
		}

		if matched {
			matches = append(matches, rule)
			if !rule.Fallthrough {
				break
			}
		}
	}

	return matches, nil
}


func MatchAll(rules []config.Rule, filename string) ([]*config.Rule, error) {
	var matches []*config.Rule

	for i := range rules {
		rule := &rules[i]
		matched, err := matchRule(rule, filename)
		if err != nil {
			return nil, err
		}

		if matched {
			matches = append(matches, rule)
		}
	}

	return matches, nil
}

func matchRule(rule *config.Rule, filename string) (bool, error) {
	// Parse URL once if needed? 
	// Actually, we can parse it inside here. It's cheap enough.
	u, err := url.Parse(filename)
	isURL := err == nil && u.Scheme != ""

	// Check OS
	if len(rule.OS) > 0 {
		matchedOS := false
		for _, osName := range rule.OS {
			if strings.ToLower(osName) == runtime.GOOS {
				matchedOS = true
				break
			}
		}
		if !matchedOS {
			return false, nil
		}
	}

	// Check Scheme
	if rule.Scheme != "" {
		if isURL && strings.EqualFold(u.Scheme, rule.Scheme) {
			return true, nil
		}
		// If scheme is specified but doesn't match, this rule is not a match
		return false, nil
	}

	// Check extensions
	if len(rule.Extensions) > 0 {
		var pathExt string
		if isURL {
			pathExt = filepath.Ext(u.Path)
		} else {
			pathExt = filepath.Ext(filename)
		}
		pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))

		for _, ruleExt := range rule.Extensions {
			if strings.ToLower(ruleExt) == pathExt {
				return true, nil
			}
		}
		// If extensions are specified but none matched, we continue to check other conditions (Regex, MIME, etc.)
		// This allows a rule to match EITHER by extension OR by regex/mime.
	}

	// Check extensions
	if len(rule.Extensions) > 0 {
		var pathExt string
		if isURL {
			pathExt = filepath.Ext(u.Path)
		} else {
			pathExt = filepath.Ext(filename)
		}
		pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))

		for _, ruleExt := range rule.Extensions {
			if strings.ToLower(ruleExt) == pathExt {
				return true, nil
			}
		}
	}

	// Check regex
	if rule.Regex != "" {
		regexMatched, err := regexp.MatchString(rule.Regex, filename)
		if err != nil {
			return false, err
		}
		if regexMatched {
			return true, nil
		}
	}

	// Check MIME type
	if rule.Mime != "" && !isURL {
		mtype, err := mimetype.DetectFile(filename)
		if err == nil {
			mimeMatched, err := regexp.MatchString(rule.Mime, mtype.String())
			if err == nil && mimeMatched {
				return true, nil
			}
		}
	}

	// Check Script
	if rule.Script != "" {
		return true, nil
	}

	return false, nil
}
