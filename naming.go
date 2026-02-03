package goorm

import (
	"regexp"
	"strings"
	"unicode"
)

// SnakeCasePlural converts a struct name to a plural snake_case table name.
// Example: User -> users, OrderItem -> order_items
//
// SnakeCasePlural 将结构体名转换为复数的 snake_case 表名。
// 示例：User -> users，OrderItem -> order_items
func SnakeCasePlural(name string) string {
	return Pluralize(SnakeCase(name))
}

// SnakeCaseSingular converts a struct name to a singular snake_case table name.
// Example: User -> user, OrderItem -> order_item
//
// SnakeCaseSingular 将结构体名转换为单数的 snake_case 表名。
// 示例：User -> user，OrderItem -> order_item
func SnakeCaseSingular(name string) string {
	return SnakeCase(name)
}

// SnakeCase converts a string from CamelCase to snake_case.
// Example: UserID -> user_id, OrderItem -> order_item
//
// SnakeCase 将字符串从 CamelCase 转换为 snake_case。
// 示例：UserID -> user_id，OrderItem -> order_item
func SnakeCase(name string) string {
	if name == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(name) + 5) // Preallocate some extra space for underscores

	for i, r := range name {
		if unicode.IsUpper(r) {
			// Add underscore before uppercase letters (except at the start)
			if i > 0 {
				// Check if the previous character was lowercase
				// or if the next character is lowercase (for acronyms like HTTP)
				prev := rune(name[i-1])
				var next rune
				if i+1 < len(name) {
					next = rune(name[i+1])
				}

				if unicode.IsLower(prev) || (unicode.IsUpper(prev) && unicode.IsLower(next)) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// CamelCase converts a string from snake_case to CamelCase.
// Example: user_id -> UserId, order_item -> OrderItem
//
// CamelCase 将字符串从 snake_case 转换为 CamelCase。
// 示例：user_id -> UserId，order_item -> OrderItem
func CamelCase(name string) string {
	if name == "" {
		return ""
	}

	var result strings.Builder
	result.Grow(len(name))

	capitalizeNext := true
	for _, r := range name {
		if r == '_' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// LowerCamelCase converts a string from snake_case to lowerCamelCase.
// Example: user_id -> userId, order_item -> orderItem
//
// LowerCamelCase 将字符串从 snake_case 转换为 lowerCamelCase。
// 示例：user_id -> userId，order_item -> orderItem
func LowerCamelCase(name string) string {
	camel := CamelCase(name)
	if len(camel) == 0 {
		return ""
	}
	return strings.ToLower(string(camel[0])) + camel[1:]
}

// Pluralize converts a singular word to its plural form.
// This is a simple implementation covering common English patterns.
//
// Pluralize 将单数词转换为复数形式。
// 这是一个简单的实现，涵盖常见的英语模式。
func Pluralize(word string) string {
	if word == "" {
		return ""
	}

	// Irregular plurals
	// 不规则复数
	irregulars := map[string]string{
		"person":   "people",
		"child":    "children",
		"man":      "men",
		"woman":    "women",
		"foot":     "feet",
		"tooth":    "teeth",
		"goose":    "geese",
		"mouse":    "mice",
		"ox":       "oxen",
		"datum":    "data",
		"index":    "indices",
		"matrix":   "matrices",
		"vertex":   "vertices",
		"analysis": "analyses",
		"crisis":   "crises",
	}

	lower := strings.ToLower(word)
	if plural, ok := irregulars[lower]; ok {
		// Preserve original case pattern
		if unicode.IsUpper(rune(word[0])) {
			return strings.ToUpper(string(plural[0])) + plural[1:]
		}
		return plural
	}

	// Words ending in -s, -x, -z, -ch, -sh -> add -es
	// 以 -s、-x、-z、-ch、-sh 结尾的词 -> 加 -es
	if strings.HasSuffix(lower, "s") ||
		strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") ||
		strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh") {
		return word + "es"
	}

	// Words ending in consonant + y -> change y to ies
	// 以辅音 + y 结尾的词 -> 将 y 改为 ies
	if len(word) > 1 && strings.HasSuffix(lower, "y") {
		lastButOne := rune(lower[len(lower)-2])
		if !isVowel(lastButOne) {
			return word[:len(word)-1] + "ies"
		}
	}

	// Words ending in -f or -fe -> change to -ves
	// 以 -f 或 -fe 结尾的词 -> 改为 -ves
	fePatterns := []string{"fe", "lf", "af", "rf", "of", "if", "ef"}
	for _, pattern := range fePatterns {
		if strings.HasSuffix(lower, pattern) {
			if strings.HasSuffix(lower, "fe") {
				return word[:len(word)-2] + "ves"
			}
			return word[:len(word)-1] + "ves"
		}
	}

	// Words ending in -o -> add -es (for common ones)
	// 以 -o 结尾的词 -> 加 -es（常见单词）
	oesToAdd := regexp.MustCompile(`(?i)(hero|potato|tomato|echo|torpedo|veto)$`)
	if oesToAdd.MatchString(word) {
		return word + "es"
	}

	// Default: add -s
	// 默认：加 -s
	return word + "s"
}

// Singularize converts a plural word to its singular form.
// This is a simple implementation covering common English patterns.
//
// Singularize 将复数词转换为单数形式。
// 这是一个简单的实现，涵盖常见的英语模式。
func Singularize(word string) string {
	if word == "" {
		return ""
	}

	// Irregular plurals (reverse lookup)
	// 不规则复数（反向查找）
	irregulars := map[string]string{
		"people":   "person",
		"children": "child",
		"men":      "man",
		"women":    "woman",
		"feet":     "foot",
		"teeth":    "tooth",
		"geese":    "goose",
		"mice":     "mouse",
		"oxen":     "ox",
		"data":     "datum",
		"indices":  "index",
		"matrices": "matrix",
		"vertices": "vertex",
		"analyses": "analysis",
		"crises":   "crisis",
	}

	lower := strings.ToLower(word)
	if singular, ok := irregulars[lower]; ok {
		if unicode.IsUpper(rune(word[0])) {
			return strings.ToUpper(string(singular[0])) + singular[1:]
		}
		return singular
	}

	// Words ending in -ies -> change to -y
	// 以 -ies 结尾的词 -> 改为 -y
	if strings.HasSuffix(lower, "ies") && len(word) > 3 {
		return word[:len(word)-3] + "y"
	}

	// Words ending in -ves -> change to -f or -fe
	// 以 -ves 结尾的词 -> 改为 -f 或 -fe
	if strings.HasSuffix(lower, "ves") {
		// Check common -fe words
		base := word[:len(word)-3]
		if strings.HasSuffix(strings.ToLower(base), "wi") || strings.HasSuffix(strings.ToLower(base), "li") {
			return base + "fe"
		}
		return base + "f"
	}

	// Words ending in -es -> remove -es
	// 以 -es 结尾的词 -> 去掉 -es
	if strings.HasSuffix(lower, "es") && len(word) > 2 {
		base := word[:len(word)-2]
		baseLower := strings.ToLower(base)
		// Check if base ends in s, x, z, ch, sh
		if strings.HasSuffix(baseLower, "s") ||
			strings.HasSuffix(baseLower, "x") ||
			strings.HasSuffix(baseLower, "z") ||
			strings.HasSuffix(baseLower, "ch") ||
			strings.HasSuffix(baseLower, "sh") {
			return base
		}
	}

	// Words ending in -s -> remove -s
	// 以 -s 结尾的词 -> 去掉 -s
	if strings.HasSuffix(lower, "s") && len(word) > 1 {
		return word[:len(word)-1]
	}

	return word
}

// isVowel checks if a rune is a vowel.
// isVowel 检查一个字符是否是元音。
func isVowel(r rune) bool {
	r = unicode.ToLower(r)
	return r == 'a' || r == 'e' || r == 'i' || r == 'o' || r == 'u'
}
