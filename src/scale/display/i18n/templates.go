package i18n

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TemplateVars represents variables that can be substituted in templates
type TemplateVars map[string]interface{}

// ProcessTemplate processes a template string with variable substitution
func ProcessTemplate(template string, vars TemplateVars) string {
	if vars == nil {
		return template
	}
	
	result := template
	
	// Regular expression to find placeholders like {variable} or {variable:format}
	re := regexp.MustCompile(`\{([^}:]+)(?::([^}]+))?\}`)
	
	// Find all matches
	matches := re.FindAllStringSubmatch(result, -1)
	
	for _, match := range matches {
		placeholder := match[0]  // Full match like {count} or {minutes:02d}
		varName := match[1]      // Variable name like "count" or "minutes"
		format := ""
		if len(match) > 2 {
			format = match[2]    // Format specifier like "02d"
		}
		
		// Get the value from vars
		if value, exists := vars[varName]; exists {
			var replacement string
			
			// Handle different types and formats
			switch v := value.(type) {
			case int:
				if format != "" {
					replacement = formatIntWithSpecifier(v, format)
				} else {
					replacement = strconv.Itoa(v)
				}
			case float64:
				if format != "" {
					replacement = fmt.Sprintf("%."+format, v)
				} else {
					replacement = fmt.Sprintf("%.1f", v)
				}
			case string:
				replacement = v
			default:
				replacement = fmt.Sprintf("%v", v)
			}
			
			result = strings.ReplaceAll(result, placeholder, replacement)
		}
	}
	
	return result
}

// formatIntWithSpecifier formats an integer with a format specifier like "02d"
func formatIntWithSpecifier(value int, format string) string {
	// Handle format specifiers like "02d"
	if strings.HasSuffix(format, "d") {
		return fmt.Sprintf("%"+format, value)
	}
	
	// Default to string conversion if format is not recognized
	return strconv.Itoa(value)
}

// CreateTemplateVars is a helper function to create TemplateVars easily
func CreateTemplateVars() TemplateVars {
	return make(TemplateVars)
}

// Set adds a variable to the template vars
func (tv TemplateVars) Set(key string, value interface{}) TemplateVars {
	tv[key] = value
	return tv
}

// SetInt is a convenience method for setting integer values
func (tv TemplateVars) SetInt(key string, value int) TemplateVars {
	tv[key] = value
	return tv
}

// SetString is a convenience method for setting string values
func (tv TemplateVars) SetString(key string, value string) TemplateVars {
	tv[key] = value
	return tv
}

// SetFloat is a convenience method for setting float values
func (tv TemplateVars) SetFloat(key string, value float64) TemplateVars {
	tv[key] = value
	return tv
}