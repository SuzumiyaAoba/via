package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/charmbracelet/lipgloss"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

type ruleResult struct {
	num        string
	name       string
	conditions []string
	result     string
	matched    bool
}

func handleExplain(cmd *cobra.Command, cfg *config.Config, filename string) error {
	// Define styles
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(0, 2).
		MarginTop(1).
		MarginBottom(1)
	
	sectionTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14")).
		Padding(0, 1).
		MarginTop(1)
	
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Width(12)
	
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))
	
	matchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)
	
	skipStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)
	
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))
	
	// Title
	fmt.Fprintln(cmd.OutOrStdout(), titleStyle.Render("═══ EXPLAIN MODE ═══"))
	fmt.Fprintln(cmd.OutOrStdout(), "")
	
	// Check if file/URL exists
	isURLType := isURL(filename)
	
	// File Information Section
	fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("FILE INFORMATION"))
	fmt.Fprintln(cmd.OutOrStdout(), "")
	
	fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Path:")+" "+valueStyle.Render(filename))
	
	if isURLType {
		fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Type:")+" "+valueStyle.Render("URL"))
		u, _ := url.Parse(filename)
		fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Scheme:")+" "+valueStyle.Render(u.Scheme))
		if u.Path != "" {
			ext := filepath.Ext(u.Path)
			if ext != "" {
				fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Extension:")+" "+valueStyle.Render(strings.TrimPrefix(ext, ".")))
			}
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Type:")+" "+valueStyle.Render("File"))
		ext := filepath.Ext(filename)
		if ext != "" {
			fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Extension:")+" "+valueStyle.Render(strings.TrimPrefix(ext, ".")))
		}
		
		// Check MIME type
		if _, err := os.Stat(filename); err == nil {
			mtype, err := mimetype.DetectFile(filename)
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  MIME Type:")+" "+valueStyle.Render(mtype.String()))
			}
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), labelStyle.Render("  Status:")+" "+errorStyle.Render("Does not exist"))
		}
	}
	
	// Rule Evaluation Section
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("RULE EVALUATION"))
	fmt.Fprintln(cmd.OutOrStdout(), "")
	
	// Prepare table data
	var results []ruleResult
	matched := false
	
	for i, rule := range cfg.Rules {
		result := ruleResult{
			num:  fmt.Sprintf("%d", i+1),
			name: rule.Name,
		}
		
		if result.name == "" {
			result.name = "-"
		}
		
		// Check each condition
		ruleMatched := false
		var matchReasons []string
		
		// OS check
		if len(rule.OS) > 0 {
			osMatch := false
			for _, osName := range rule.OS {
				if strings.ToLower(osName) == runtime.GOOS {
					osMatch = true
					break
				}
			}
			if osMatch {
				result.conditions = append(result.conditions, "✓ OS: "+runtime.GOOS)
				matchReasons = append(matchReasons, "OS")
			} else {
				result.conditions = append(result.conditions, "✗ OS mismatch")
				result.result = errorStyle.Render("SKIP")
				results = append(results, result)
				continue
			}
		}
		
		// Scheme check
		if rule.Scheme != "" {
			u, _ := url.Parse(filename)
			if isURLType && strings.ToLower(u.Scheme) == strings.ToLower(rule.Scheme) {
				result.conditions = append(result.conditions, "✓ Scheme: "+rule.Scheme)
				matchReasons = append(matchReasons, "Scheme")
				ruleMatched = true
			} else {
				result.conditions = append(result.conditions, "✗ Scheme: "+rule.Scheme)
				result.result = errorStyle.Render("SKIP")
				results = append(results, result)
				continue
			}
		}
		
		// Extension check
		if !ruleMatched && len(rule.Extensions) > 0 {
			var pathExt string
			if isURLType {
				u, _ := url.Parse(filename)
				pathExt = filepath.Ext(u.Path)
			} else {
				pathExt = filepath.Ext(filename)
			}
			pathExt = strings.ToLower(strings.TrimPrefix(pathExt, "."))
			
			extMatched := false
			for _, ruleExt := range rule.Extensions {
				if strings.ToLower(ruleExt) == pathExt {
					result.conditions = append(result.conditions, "✓ Ext: ."+pathExt)
					matchReasons = append(matchReasons, "Extension")
					ruleMatched = true
					extMatched = true
					break
				}
			}
			if !extMatched {
				result.conditions = append(result.conditions, skipStyle.Render("○ Ext: "+fmt.Sprintf("%v", rule.Extensions)))
			}
		}
		
		// Regex check
		if !ruleMatched && rule.Regex != "" {
			regexMatched, _ := regexp.MatchString(rule.Regex, filename)
			if regexMatched {
				result.conditions = append(result.conditions, "✓ Regex")
				matchReasons = append(matchReasons, "Regex")
				ruleMatched = true
			} else {
				result.conditions = append(result.conditions, skipStyle.Render("○ Regex: "+rule.Regex))
			}
		}
		
		// MIME check
		if !ruleMatched && rule.Mime != "" && !isURLType {
			if _, err := os.Stat(filename); err == nil {
				mtype, err := mimetype.DetectFile(filename)
				if err == nil {
					mimeMatched, _ := regexp.MatchString(rule.Mime, mtype.String())
					if mimeMatched {
						result.conditions = append(result.conditions, "✓ MIME")
						matchReasons = append(matchReasons, "MIME")
						ruleMatched = true
					} else {
						result.conditions = append(result.conditions, skipStyle.Render("○ MIME: "+rule.Mime))
					}
				}
			}
		}
		
		if ruleMatched {
			result.result = matchStyle.Render("[MATCH]")
			if rule.Fallthrough {
				result.result += skipStyle.Render(" →")
			}
			result.matched = true
			matched = true
			results = append(results, result)
			if !rule.Fallthrough {
				break
			}
		} else {
			result.result = skipStyle.Render("—")
			results = append(results, result)
		}
	}
	
	// Prepare table rows
	rows := make([][]string, len(results))
	for i, r := range results {
		rows[i] = []string{
			r.num,
			r.name,
			strings.Join(r.conditions, "\n"),
			r.result,
		}
	}

	// Render table
	headers := []string{"#", "Rule Name", "Conditions", "Result"}
	fmt.Fprintln(cmd.OutOrStdout(), createStyledTable(headers, rows))
	
	// Result Section
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), sectionTitleStyle.Render("RESULT"))
	fmt.Fprintln(cmd.OutOrStdout(), "")
	
	if !matched {
		fmt.Fprintln(cmd.OutOrStdout(), "  "+skipStyle.Render("No rules matched."))
		fmt.Fprintln(cmd.OutOrStdout(), "")
		if cfg.DefaultCommand != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Action:")+" "+valueStyle.Render("Execute default command"))
			fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Command:")+" "+valueStyle.Render(cfg.DefaultCommand))
		} else {
			fmt.Fprintln(cmd.OutOrStdout(), "  "+labelStyle.Render("Action:")+" "+valueStyle.Render("Use system default application"))
		}
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "  "+matchStyle.Render("[OK] Rule matched successfully"))
	}
	
	fmt.Fprintln(cmd.OutOrStdout(), "")
	
	return nil
}
