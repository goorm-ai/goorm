package goorm

import "testing"

// TestSnakeCase tests the SnakeCase function.
// TestSnakeCase 测试 SnakeCase 函数。
func TestSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserID", "user_id"},
		{"OrderItem", "order_item"},
		{"HTTPServer", "http_server"},
		{"ID", "id"},
		{"CreatedAt", "created_at"},
		{"IsVIP", "is_vip"},
		{"XMLParser", "xml_parser"},
		{"IOReader", "io_reader"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCamelCase tests the CamelCase function.
// TestCamelCase 测试 CamelCase 函数。
func TestCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_id", "UserId"},
		{"order_item", "OrderItem"},
		{"created_at", "CreatedAt"},
		{"is_vip", "IsVip"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := CamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("CamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPluralize tests the Pluralize function.
// TestPluralize 测试 Pluralize 函数。
func TestPluralize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "users"},
		{"category", "categories"},
		{"box", "boxes"},
		{"bus", "buses"},
		{"knife", "knives"},
		{"leaf", "leaves"},
		{"person", "people"},
		{"child", "children"},
		{"man", "men"},
		{"hero", "heroes"},
		{"photo", "photos"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Pluralize(tt.input)
			if result != tt.expected {
				t.Errorf("Pluralize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSingularize tests the Singularize function.
// TestSingularize 测试 Singularize 函数。
func TestSingularize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "user"},
		{"categories", "category"},
		{"boxes", "box"},
		{"people", "person"},
		{"children", "child"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Singularize(tt.input)
			if result != tt.expected {
				t.Errorf("Singularize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSnakeCasePlural tests the SnakeCasePlural function.
// TestSnakeCasePlural 测试 SnakeCasePlural 函数。
func TestSnakeCasePlural(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "users"},
		{"OrderItem", "order_items"},
		{"UserProfile", "user_profiles"},
		{"Person", "people"},
		{"Category", "categories"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SnakeCasePlural(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeCasePlural(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
