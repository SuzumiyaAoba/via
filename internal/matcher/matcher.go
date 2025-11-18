package matcher

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/gabriel-vasile/mimetype"
)

func Match(rules []config.Rule, defaultCommand string, filename string) (*config.Rule, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	// Remove dot from extension if present for comparison, or keep it?
	// Usually config might have ".txt" or "txt". Let's handle both or assume one.
	// Let's assume config has "txt" (no dot) for simplicity, or handle both.
	// Better: trim dot from file ext.
	ext = strings.TrimPrefix(ext, ".")

	for _, rule := range rules {
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

		// Check extensions
		for _, ruleExt := range rule.Extensions {
			if strings.ToLower(ruleExt) == ext {
				return &rule, nil
			}
		}

		// Check regex
		if rule.Regex != "" {
			matched, err := regexp.MatchString(rule.Regex, filename)
			if err != nil {
				return nil, err
			}
			if matched {
				return &rule, nil
			}
		}

		// Check MIME type
		if rule.Mime != "" {
			mtype, err := mimetype.DetectFile(filename)
			if err != nil {
				// If file cannot be read, ignore MIME match? Or return error?
				// For now, ignore and continue to next rule.
				continue
			}
			// Use regex for MIME match? Or exact match?
			// Let's use regex for flexibility (e.g. "image/.*").
			matched, err := regexp.MatchString(rule.Mime, mtype.String())
			if err != nil {
				return nil, err
			}
			if matched {
				return &rule, nil
			}
		}
	}

	if defaultCommand != "" {
		return &config.Rule{Command: defaultCommand}, nil
	}

	return nil, nil
}
