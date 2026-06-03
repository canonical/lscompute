package types

import (
	"encoding/json"
	"testing"

	"go.yaml.in/yaml/v4"
)

func TestHexIntMarshalJSON(t *testing.T) {
	tests := []struct {
		val  HexInt
		want string
	}{
		{0x8086, `"0x8086"`},
		{0x0, `"0x0"`},
		{0x10DE, `"0x10DE"`},
		{0x1234, `"0x1234"`},
	}
	for _, tc := range tests {
		data, err := json.Marshal(tc.val)
		if err != nil {
			t.Errorf("Marshal(%#x) error: %v", int(tc.val), err)
			continue
		}
		if string(data) != tc.want {
			t.Errorf("Marshal(%#x) = %s, want %s", int(tc.val), data, tc.want)
		}
	}
}

func TestHexIntUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HexInt
		wantErr bool
	}{
		{"0x prefix lowercase", `"0x8086"`, 0x8086, false},
		{"no prefix", `"8086"`, 0x8086, false},
		{"uppercase hex digits", `"10DE"`, 0x10DE, false},
		{"zero", `"0x0"`, 0, false},
		{"empty string → zero", `""`, 0, false},
		{"invalid hex", `"ZZZZ"`, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var h HexInt
			err := json.Unmarshal([]byte(tc.input), &h)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UnmarshalJSON(%s): expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s): unexpected error: %v", tc.input, err)
			}
			if h != tc.want {
				t.Errorf("UnmarshalJSON(%s) = 0x%x, want 0x%x", tc.input, int(h), int(tc.want))
			}
		})
	}
}

// TestHexIntJSONRoundTrip verifies that Marshal → Unmarshal is identity.
func TestHexIntJSONRoundTrip(t *testing.T) {
	vals := []HexInt{0, 0x1, 0x8086, 0x10DE, 0xFFFF}
	for _, v := range vals {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(0x%x): %v", int(v), err)
		}
		var got HexInt
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got != v {
			t.Errorf("round-trip 0x%x: got 0x%x", int(v), int(got))
		}
	}
}

func TestHexIntMarshalYAML(t *testing.T) {
	tests := []struct {
		val  HexInt
		want string
	}{
		{0x8086, "0x8086"},
		{0x0, "0x0"},
		{0x10DE, "0x10DE"},
	}
	for _, tc := range tests {
		val, err := tc.val.MarshalYAML()
		if err != nil {
			t.Errorf("MarshalYAML(%#x) error: %v", int(tc.val), err)
			continue
		}
		got, ok := val.(string)
		if !ok {
			t.Errorf("MarshalYAML(%#x) returned %T, want string", int(tc.val), val)
			continue
		}
		if got != tc.want {
			t.Errorf("MarshalYAML(%#x) = %q, want %q", int(tc.val), got, tc.want)
		}
	}
}

func TestHexIntUnmarshalYAML(t *testing.T) {
	type wrapper struct {
		Val HexInt `yaml:"val"`
	}

	tests := []struct {
		name    string
		yaml    string
		want    HexInt
		wantErr bool
	}{
		{"0x prefix", "val: \"0x8086\"\n", 0x8086, false},
		{"no prefix", "val: \"8086\"\n", 0x8086, false},
		{"uppercase", "val: \"10DE\"\n", 0x10DE, false},
		{"empty string → zero", "val: \"\"\n", 0, false},
		{"invalid hex", "val: \"ZZZZ\"\n", 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var w wrapper
			err := yaml.Unmarshal([]byte(tc.yaml), &w)
			if tc.wantErr {
				if err == nil {
					t.Errorf("UnmarshalYAML(%q): expected error, got nil", tc.yaml)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalYAML(%q): unexpected error: %v", tc.yaml, err)
			}
			if w.Val != tc.want {
				t.Errorf("UnmarshalYAML(%q) = 0x%x, want 0x%x", tc.yaml, int(w.Val), int(tc.want))
			}
		})
	}
}

// TestHexIntYAMLRoundTrip verifies that MarshalYAML → UnmarshalYAML is identity.
func TestHexIntYAMLRoundTrip(t *testing.T) {
	type wrapper struct {
		Val HexInt `yaml:"val"`
	}

	vals := []HexInt{0, 0x1, 0x8086, 0x10DE, 0xFFFF}
	for _, v := range vals {
		w := wrapper{Val: v}
		data, err := yaml.Marshal(w)
		if err != nil {
			t.Fatalf("Marshal(0x%x): %v", int(v), err)
		}
		var got wrapper
		if err := yaml.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got.Val != v {
			t.Errorf("round-trip 0x%x: got 0x%x", int(v), int(got.Val))
		}
	}
}
