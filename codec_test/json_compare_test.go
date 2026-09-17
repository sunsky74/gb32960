package codec_test

import "testing"

func TestJsonSemanticEqual_Identical(t *testing.T) {
	if err := jsonSemanticEqual([]byte(`{"a":1}`), []byte(`{"a":1}`)); err != nil {
		t.Errorf("identical JSON should be equal: %v", err)
	}
}

func TestJsonSemanticEqual_FloatTolerance(t *testing.T) {
	// Go 的 float64 50.0 与 Java 的 BigDecimal "50.0000000001",应在 1e-6 内匹配
	if err := jsonSemanticEqual([]byte(`{"v":50.0}`), []byte(`{"v":50.0000000001}`)); err != nil {
		t.Errorf("float within tolerance should be equal: %v", err)
	}
}

func TestJsonSemanticEqual_MissingKey(t *testing.T) {
	// Go 有 Java 没有的额外键
	if err := jsonSemanticEqual([]byte(`{"a":1}`), []byte(`{"a":1,"b":2}`)); err == nil {
		t.Error("extra key in Go should cause mismatch")
	}
}

func TestJsonSemanticEqual_MissingKeyReverse(t *testing.T) {
	// Java 有 Go 没有的键
	if err := jsonSemanticEqual([]byte(`{"a":1,"b":2}`), []byte(`{"a":1}`)); err == nil {
		t.Error("missing key in Go should cause mismatch")
	}
}
