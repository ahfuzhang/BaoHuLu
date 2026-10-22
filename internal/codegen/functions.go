package codegen

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/ahfuzhang/BaoHuLu/internal/schema"
)

func genericFunctions() template.FuncMap {
	return template.FuncMap{
		"matches":  regexp.MatchString,
		"base":     filepath.Base,
		"ext":      filepath.Ext,
		"layout":   memoryLayout,
		"tagBytes": schema.TagBytes,
		"reachable": func(graph map[string][]string, start, target string) bool {
			visited := map[string]bool{}
			var visit func(string) bool
			visit = func(current string) bool {
				if current == target {
					return true
				}
				if visited[current] {
					return false
				}
				visited[current] = true
				for _, next := range graph[current] {
					if visit(next) {
						return true
					}
				}
				return false
			}
			return visit(start)
		},
		"set":    func(target map[string]any, key string, value any) string { target[key] = value; return "" },
		"get":    func(target map[string]any, key string) any { return target[key] },
		"has":    func(target map[string]any, key string) bool { _, exists := target[key]; return exists },
		"append": func(target []any, values ...any) []any { return append(target, values...) },
		"concat": func(parts ...string) string { return strings.Join(parts, "") },
		"add":    func(left, right int) int { return left + right },
		"sub":    func(left, right int) int { return left - right },
		"mul":    func(left, right int) int { return left * right },
		"div": func(left, right int) (int, error) {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		},
		"mod": func(left, right int) (int, error) {
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return left % right, nil
		},
		"atoi":      strconv.Atoi,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"contains":  strings.Contains,
		"trim":      strings.TrimSpace,
		"upperFirst": func(value string) string {
			characters := []rune(value)
			if len(characters) > 0 {
				characters[0] = unicode.ToUpper(characters[0])
			}
			return string(characters)
		},
		"padRight": func(value string, width int) string {
			if width <= len(value) {
				return value
			}
			return value + strings.Repeat(" ", width-len(value))
		},
		"reverse": func(value any) ([]any, error) {
			items := reflect.ValueOf(value)
			if items.Kind() != reflect.Slice && items.Kind() != reflect.Array {
				return nil, fmt.Errorf("reverse requires a sequence")
			}
			result := make([]any, items.Len())
			for index := range result {
				result[len(result)-1-index] = items.Index(index).Interface()
			}
			return result, nil
		},
		"sortBy": func(values []any, key string) ([]any, error) {
			result := append([]any(nil), values...)
			for _, value := range result {
				fields, ok := value.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("sortBy requires dictionaries")
				}
				if _, ok := fields[key].(int); !ok {
					return nil, fmt.Errorf("sortBy key %s must be an integer", key)
				}
			}
			sort.SliceStable(result, func(left, right int) bool {
				return result[left].(map[string]any)[key].(int) < result[right].(map[string]any)[key].(int)
			})
			return result, nil
		},
	}
}
