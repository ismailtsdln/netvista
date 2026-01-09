package screenshot

import (
	"testing"
)

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		h1       string
		h2       string
		expected int
	}{
		{"p:0000000000000000", "p:0000000000000000", 0},
		{"p:ffffffffffffffff", "p:ffffffffffffffff", 0},
		{"p:0000000000000000", "p:0000000000000001", 1},
		{"p:ffffffffffffffff", "p:0000000000000000", 64},
	}

	for _, tt := range tests {
		dist, err := HammingDistance(tt.h1, tt.h2)
		if err != nil && tt.expected != -1 { // Assuming -1 means error expected if I added such cases
			t.Errorf("HammingDistance(%s, %s) returned error: %v", tt.h1, tt.h2, err)
			continue
		}
		if dist != tt.expected {
			t.Errorf("HammingDistance(%s, %s) = %d; want %d", tt.h1, tt.h2, dist, tt.expected)
		}
	}

}
