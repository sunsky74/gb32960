package utils

import (
	"math"
	"testing"
)

func TestWGS84ToGCJ02_Tiananmen(t *testing.T) {
	c := WGS84ToGCJ02(116.397428, 39.90923)
	if math.Abs(c.Longitude-116.403668) > 0.0001 {
		t.Errorf("longitude: expected ~116.403668, got %f", c.Longitude)
	}
	if math.Abs(c.Latitude-39.910627) > 0.0001 {
		t.Errorf("latitude: expected ~39.910627, got %f", c.Latitude)
	}
}

func TestWGS84ToGCJ02_OutsideChina(t *testing.T) {
	// 东京:应原样返回
	c := WGS84ToGCJ02(139.6917, 35.6895)
	if c.Longitude != 139.6917 || c.Latitude != 35.6895 {
		t.Error("coordinates outside China should not be converted")
	}
}

func TestWGS84ToGCJ02_Shanghai(t *testing.T) {
	c := WGS84ToGCJ02(121.473701, 31.230416)
	if math.Abs(c.Longitude-121.478220) > 0.0001 {
		t.Errorf("longitude: expected ~121.478220, got %f", c.Longitude)
	}
	if math.Abs(c.Latitude-31.228466) > 0.0001 {
		t.Errorf("latitude: expected ~31.228466, got %f", c.Latitude)
	}
}
