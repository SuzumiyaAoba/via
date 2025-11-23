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
		// If extensions are specified but none matched, check other conditions?
		// No, usually extensions are a primary match condition.
		// But wait, the original logic was:
		// if !matched && len(rule.Extensions) > 0 { ... }
		// It was OR logic between conditions?
		// No, it was sequential checks.
		// Let's re-read original logic carefully.
		
		// Original Match:
		// 1. Check OS -> if fail, continue (AND)
		// 2. Check Scheme -> if set: if match, matched=true. else continue (AND)
		// 3. Check Extensions -> if !matched && set: if match, matched=true.
		// 4. Check Regex -> if !matched && set: if match, matched=true.
		// 5. Check MIME -> if !matched && set: if match, matched=true.
		// 6. Check Script -> if !matched && set: matched=true.
		
		// So it is:
		// OS (AND)
		// (Scheme OR Extensions OR Regex OR MIME OR Script)
		
		// But Scheme check was special: if set and not match, it skips the rule.
		// So Scheme is also AND if set?
		// "If scheme is specified but doesn't match, skip this rule" -> Yes, AND.
		
		// So:
		// OS (AND)
		// Scheme (AND if set)
		// (Extensions OR Regex OR MIME OR Script)
		
		// Wait, if Scheme matches, it sets matched=true.
		// And subsequent checks are `if !matched`.
		// So if Scheme matches, we are done? Yes.
		
		// So:
		// If OS fails -> return false
		// If Scheme set:
		//    If match -> return true
		//    If no match -> return false
		
		// If Extensions set:
		//    If match -> return true
		//    If no match -> continue to next check (Regex)
		
		// Wait, if Extensions set and NO match, do we fail?
		// Original code:
		// if !matched && len(rule.Extensions) > 0 { ... if match { matched = true } }
		// It doesn't say "else return false".
		// So if extensions don't match, it falls through to Regex.
		
		// BUT, usually if you specify extensions, you expect one of them to match.
		// However, the code allows a rule with Extensions AND Regex.
		// If extensions don't match, maybe regex does?
		// Example: ext: [txt], regex: .*foo
		// file: foo (no ext). Extensions check fails. Regex check passes.
		// Should it match?
		// The original code allows this.
		
		// So my refactoring must preserve this "OR" behavior for the content checks.
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
