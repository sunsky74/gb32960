package codec_test

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

// jsonSemanticEqual 对两个 JSON 字节切片做语义相等比较。
// 数值使用 InDelta 比较(容差 1e-6),以处理
// Go float64 与 Java BigDecimal 的序列化差异。
// 同时执行正向(Java→Go)和反向(Go→Java)键检查,
// 以捕获 Go 可能多出而 Java 没有的字段。
func jsonSemanticEqual(javaJSON, goJSON []byte) error {
	var javaVal, goVal interface{}
	if err := json.Unmarshal(javaJSON, &javaVal); err != nil {
		return fmt.Errorf("unmarshal java json: %w", err)
	}
	if err := json.Unmarshal(goJSON, &goVal); err != nil {
		return fmt.Errorf("unmarshal go json: %w", err)
	}
	return compareValues(javaVal, goVal, "")
}

const floatTolerance = 1e-6

func compareValues(java, goVal interface{}, path string) error {
	jv := reflect.ValueOf(java)
	gv := reflect.ValueOf(goVal)

	// 处理 nil
	if !jv.IsValid() && !gv.IsValid() {
		return nil
	}
	if !jv.IsValid() || !gv.IsValid() {
		return fmt.Errorf("%s: nil mismatch (java=%v, go=%v)", path, java, goVal)
	}

	// 解包 interface
	if jv.Kind() == reflect.Interface {
		jv = jv.Elem()
	}
	if gv.Kind() == reflect.Interface {
		gv = gv.Elem()
	}

	// 数值类型用差值比较处理
	if isNumeric(jv) && isNumeric(gv) {
		jf := toFloat64(jv)
		gf := toFloat64(gv)
		if math.IsNaN(jf) && math.IsNaN(gf) {
			return nil
		}
		if math.Abs(jf-gf) > floatTolerance {
			return fmt.Errorf("%s: numeric mismatch (java=%v, go=%v, delta=%g)",
				path, jf, gf, math.Abs(jf-gf))
		}
		return nil
	}

	// 处理数组/切片
	if jv.Kind() == reflect.Slice && gv.Kind() == reflect.Slice {
		if jv.Len() != gv.Len() {
			return fmt.Errorf("%s: length mismatch (java=%d, go=%d)", path, jv.Len(), gv.Len())
		}
		for i := 0; i < jv.Len(); i++ {
			if err := compareValues(jv.Index(i).Interface(), gv.Index(i).Interface(),
				fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		return nil
	}

	// 处理 map/对象:双向键检查
	if jv.Kind() == reflect.Map && gv.Kind() == reflect.Map {
		// 正向遍历:每个 Java 键都必须存在于 Go 中
		for _, k := range jv.MapKeys() {
			jItem := jv.MapIndex(k).Interface()
			gItem := gv.MapIndex(k)
			if !gItem.IsValid() {
				return fmt.Errorf("%s.%v: missing key in Go output", path, k)
			}
			if err := compareValues(jItem, gItem.Interface(),
				fmt.Sprintf("%s.%v", path, k)); err != nil {
				return err
			}
		}
		// 反向遍历:每个 Go 键都必须存在于 Java 中
		for _, k := range gv.MapKeys() {
			if !jv.MapIndex(k).IsValid() {
				return fmt.Errorf("%s.%v: unexpected key in Go output (absent in Java)", path, k)
			}
		}
		return nil
	}

	// 兜底:精确相等
	if !reflect.DeepEqual(java, goVal) {
		return fmt.Errorf("%s: value mismatch (java=%v, go=%v)", path, java, goVal)
	}
	return nil
}

func isNumeric(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

func toFloat64(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	}
	return 0
}
