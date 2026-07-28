package resources

import (
	"encoding/json"
	"testing"
)

// APIRequest deliberately OMITS empty values rather than sending zero-valued
// fields, so that an unset Terraform attribute never overwrites a server-side
// value. These tests pin that omission contract.

func TestNewAPIRequest_PresetsTokenOnlyWhenSet(t *testing.T) {
	withToken := NewAkeylessClient("https://api.example.invalid", "tok-123").NewAPIRequest()
	if got, ok := withToken.fields["token"]; !ok || got != "tok-123" {
		t.Fatalf("expected token pre-set to %q, got %v (present=%v)", "tok-123", got, ok)
	}

	withoutToken := NewAkeylessClient("https://api.example.invalid", "").NewAPIRequest()
	if _, ok := withoutToken.fields["token"]; ok {
		t.Fatalf("expected no token field when client token is empty, got %v", withoutToken.fields)
	}
}

func TestAPIRequest_OmitsEmptyValues(t *testing.T) {
	r := NewAkeylessClient("https://api.example.invalid", "").NewAPIRequest()

	r.SetString("kept-string", "value")
	r.SetString("dropped-string", "")
	r.SetStringSlice("kept-slice", []string{"a", "b"})
	r.SetStringSlice("dropped-slice", nil)
	r.SetStringSlice("dropped-empty-slice", []string{})
	r.Set("dropped-nil", nil)
	// Booleans and ints are always sent: false and 0 are meaningful values.
	r.SetBool("kept-false", false)
	r.SetInt64("kept-zero", 0)

	for _, key := range []string{"dropped-string", "dropped-slice", "dropped-empty-slice", "dropped-nil"} {
		if _, ok := r.fields[key]; ok {
			t.Errorf("expected %q to be omitted, but it was present", key)
		}
	}
	for _, key := range []string{"kept-string", "kept-slice", "kept-false", "kept-zero"} {
		if _, ok := r.fields[key]; !ok {
			t.Errorf("expected %q to be present, but it was omitted", key)
		}
	}
}

func TestAPIRequest_MarshalJSONEmitsFieldsAtTopLevel(t *testing.T) {
	r := NewAkeylessClient("https://api.example.invalid", "tok").NewAPIRequest()
	r.SetString("name", "my-secret")

	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var round map[string]interface{}
	if err := json.Unmarshal(raw, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if round["name"] != "my-secret" || round["token"] != "tok" {
		t.Fatalf("expected fields at top level, got %s", raw)
	}
}

func TestGetNestedString(t *testing.T) {
	data := map[string]interface{}{
		"name": "top",
		"nested": map[string]interface{}{
			"deep": map[string]interface{}{"leaf": "found"},
			"flag": true,
		},
		"whole":      float64(3),
		"fractional": float64(3.5),
		"notAMap":    "scalar",
	}

	cases := []struct {
		path    string
		want    string
		wantsOK bool
	}{
		{"name", "top", true},
		{"nested.deep.leaf", "found", true},
		{"nested.flag", "true", true},
		// A whole float64 must render as an integer, not "3e+00" — Terraform
		// state comparisons are string-based, so formatting drift is a diff.
		{"whole", "3", true},
		{"fractional", "3.5", true},
		{"missing", "", false},
		{"nested.missing", "", false},
		{"notAMap.leaf", "", false},
	}

	for _, tc := range cases {
		got, ok := GetNestedString(data, tc.path)
		if got != tc.want || ok != tc.wantsOK {
			t.Errorf("GetNestedString(%q) = (%q, %v), want (%q, %v)", tc.path, got, ok, tc.want, tc.wantsOK)
		}
	}
}

func TestGetNestedStringSlice(t *testing.T) {
	data := map[string]interface{}{
		"tags": []interface{}{"a", "b"},
		"nested": map[string]interface{}{
			"tags": []interface{}{"c"},
		},
		// Non-string members are skipped, not coerced.
		"mixed":  []interface{}{"keep", 42},
		"scalar": "not-a-slice",
	}

	if got, ok := GetNestedStringSlice(data, "tags"); !ok || len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("tags = (%v, %v), want ([a b], true)", got, ok)
	}
	if got, ok := GetNestedStringSlice(data, "nested.tags"); !ok || len(got) != 1 || got[0] != "c" {
		t.Errorf("nested.tags = (%v, %v), want ([c], true)", got, ok)
	}
	if got, ok := GetNestedStringSlice(data, "mixed"); !ok || len(got) != 1 || got[0] != "keep" {
		t.Errorf("mixed = (%v, %v), want ([keep], true)", got, ok)
	}
	if _, ok := GetNestedStringSlice(data, "scalar"); ok {
		t.Error("scalar: expected ok=false for a non-slice value")
	}
	if _, ok := GetNestedStringSlice(data, "missing"); ok {
		t.Error("missing: expected ok=false")
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"a.b.c", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", nil},
		// Empty segments collapse rather than producing "" lookups.
		{"a..b", []string{"a", "b"}},
		{".leading", []string{"leading"}},
		{"trailing.", []string{"trailing"}},
	}

	for _, tc := range cases {
		got := splitPath(tc.in)
		if len(got) != len(tc.want) {
			t.Errorf("splitPath(%q) = %v, want %v", tc.in, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitPath(%q) = %v, want %v", tc.in, got, tc.want)
				break
			}
		}
	}
}
