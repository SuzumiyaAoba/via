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
	
	// Check Scheme
	u, err := url.Parse(filename)
	isURL := err == nil && u.Scheme != ""

	for i := range rules {
		rule := &rules[i]
		matched := false
		
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
				continue
			}
		}

		// Check Scheme
		if rule.Scheme != "" {
			if isURL && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
				matched = true
			} else {
				// If scheme is specified but doesn't match, skip this rule
				continue
			}
		}

		// Check extensions
		if !matched && len(rule.Extensions) > 0 {
			var pathExt string
			if isURL {
				pathExt = filepath.Ext(u.Path)
			} else {
				pathExt = filepath.Ext(filename)
			}
			pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))

			for _, ruleExt := range rule.Extensions {
				if strings.ToLower(ruleExt) == pathExt {
					matched = true
					break
				}
			}
		}

		// Check regex
		if !matched && rule.Regex != "" {
			regexMatched, err := regexp.MatchString(rule.Regex, filename)
			if err != nil {
				return nil, err
			}
			if regexMatched {
				matched = true
			}
		}

		// Check MIME type
		if !matched && rule.Mime != "" && !isURL {
			mtype, err := mimetype.DetectFile(filename)
			if err == nil {
				mimeMatched, err := regexp.MatchString(rule.Mime, mtype.String())
				if err == nil && mimeMatched {
					matched = true
				}
			}
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

	// Check Scheme
	u, err := url.Parse(filename)
	isURL := err == nil && u.Scheme != ""

	for i := range rules {
		rule := &rules[i]
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
				continue
			}
		}

		// Check Scheme
		if rule.Scheme != "" {
			if isURL && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
				matches = append(matches, rule)
				continue
			}
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
					matches = append(matches, rule)
					goto NextRule
				}
			}
		}

		// Check regex
		if rule.Regex != "" {
			matched, err := regexp.MatchString(rule.Regex, filename)
			if err != nil {
				return nil, err
			}
			if matched {
				matches = append(matches, rule)
				continue
			}
		}

		// Check MIME type
		if rule.Mime != "" && !isURL {
			mtype, err := mimetype.DetectFile(filename)
			if err == nil {
				matched, err := regexp.MatchString(rule.Mime, mtype.String())
				if err == nil && matched {
					matches = append(matches, rule)
					continue
				}
			}
		}

	NextRule:
	}

	return matches, nil
}
