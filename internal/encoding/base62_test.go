package encoding

import (
	"testing"
)

func TestBase62Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected string
	}{
		{"zero", 0, "0"},
		{"one", 1, "1"},
		{"nine", 9, "9"},
		{"ten", 10, "a"},
		{"thirty_five", 35, "z"},
		{"thirty_six", 36, "A"},
		{"sixty_one", 61, "Z"},
		{"sixty_two", 62, "10"},
		{"sixty_three", 63, "11"},
		{"one_hundred", 100, "1C"},
		{"one_thousand", 1000, "g8"},
		{"ten_thousand", 10000, "2Bi"},
		{"hundred_thousand", 100000, "q0U"},
		{"million", 1000000, "4c92"},
		{"large_number", 123456789, "8m0Kx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base62Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Base62Encode(%d) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBase62Encode_Valid(t *testing.T) {
	tests := []struct {
		name     string
		actual   uint64
		expected string
	}{
		{"zero", 123456789, "8m0Kx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base62Encode(tt.actual)
			if result != tt.expected {
				t.Errorf("Base62Encode(%d) = %v, want %v", tt.actual, result, tt.expected)
			}
		})
	}
}

func TestBase62Encode_Consistency(t *testing.T) {
	// Test that encoding is consistent
	input := uint64(12345)
	result1 := Base62Encode(input)
	result2 := Base62Encode(input)

	if result1 != result2 {
		t.Errorf("Base62Encode(%d) returned different results: %v vs %v", input, result1, result2)
	}
}

func TestBase62Encode_UniqueValues(t *testing.T) {
	// Test that different inputs produce different outputs
	inputs := []uint64{1, 2, 3, 10, 100, 1000, 10000}
	results := make(map[string]bool)

	for _, input := range inputs {
		result := Base62Encode(input)
		if results[result] {
			t.Errorf("Base62Encode(%d) = %v, but this result was already produced", input, result)
		}
		results[result] = true
	}
}

func BenchmarkBase62Encode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Base62Encode(uint64(i))
	}
}
