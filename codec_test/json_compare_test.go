package codec_test

import "testing"

func TestJsonSemanticEqual_Identical(t *testing.T) {
	if err := jsonSemanticEqual([]byte(`{"a":1}`), []byte(`{"a":1}`)); err != nil {
		t.Errorf("identical JSON should be equal: %v", err)
	}
}

func TestJsonSemanticEqual_FloatTolerance(t *testing.T) {
	// Go float64 50.0 vs Java BigDecimal "50.0000000001" — should match within 1e-6
	if err := jsonSemanticEqual([]byte(`{"v":50.0}`), []byte(`{"v":50.0000000001}`)); err != nil {
		t.Errorf("float within tolerance should be equal: %v", err)
	}
}

func TestJsonSemanticEqual_MissingKey(t *testing.T) {
	// Go has extra key that Java doesn't
	if err := jsonSemanticEqual([]byte(`{"a":1}`), []byte(`{"a":1,"b":2}`)); err == nil {
		t.Error("extra key in Go should cause mismatch")
	}
}

func TestJsonSemanticEqual_MissingKeyReverse(t *testing.T) {
	// Java has key that Go doesn't
	if err := jsonSemanticEqual([]byte(`{"a":1,"b":2}`), []byte(`{"a":1}`)); err == nil {
		t.Error("missing key in Go should cause mismatch")
	}
}
