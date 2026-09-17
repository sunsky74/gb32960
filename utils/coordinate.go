package utils

import "math"

const (
	pi   = 3.1415926535897932384626
	axis = 6378245.0              // 克拉索夫斯基椭球长半轴
	ee   = 0.00669342162296594323 // 1 - e²(椭球偏心率平方)
)

// Coordinate 保存经度和纬度。
type Coordinate struct {
	Longitude float64
	Latitude  float64
}

// WGS84ToGCJ02 将 WGS84 坐标转换为 GCJ02(火星坐标系)。
// 中国以外的坐标原样返回。
// 算法与 Java CoordinateUtils.wgs84ToGcj02() 完全一致:
// 克拉索夫斯基椭球参数,偏移量 (lon-105, lat-35) + 椭球修正。
func WGS84ToGCJ02(wgsLon, wgsLat float64) Coordinate {
	if outOfChina(wgsLat, wgsLon) {
		return Coordinate{Longitude: wgsLon, Latitude: wgsLat}
	}

	// 第 1 步:使用偏移量 (lon-105, lat-35) 计算地理增量
	dLat := transformLat(wgsLon-105.0, wgsLat-35.0)
	dLon := transformLon(wgsLon-105.0, wgsLat-35.0)

	// 第 2 步:椭球修正(克拉索夫斯基参数)
	radLat := wgsLat / 180.0 * math.Pi
	magic := 1 - ee*math.Sin(radLat)*math.Sin(radLat)
	sqrtMagic := math.Sqrt(magic)

	// 第 3 步:带椭球修正的最终坐标
	finalLat := wgsLat + (dLat*180.0)/((axis*(1-ee))/(magic*sqrtMagic)*math.Pi)
	finalLon := wgsLon + (dLon*180.0)/(axis/sqrtMagic*math.Cos(radLat)*math.Pi)

	// 第 4 步:四舍五入到 6 位小数(与 Java BigDecimal.setScale(6, HALF_UP) 一致)
	return Coordinate{
		Longitude: math.Round(finalLon*1_000_000) / 1_000_000,
		Latitude:  math.Round(finalLat*1_000_000) / 1_000_000,
	}
}

func outOfChina(lat, lon float64) bool {
	return lon < 72.004 || lon > 137.8347 || lat < 0.8293 || lat > 55.8271
}

// transformLat:纬度偏移的多项式公式。
// 输入 x = wgsLon - 105.0,y = wgsLat - 35.0(WGS84ToGCJ02 调用时)。
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*math.Pi) + 320.0*math.Sin(y/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}

// transformLon:经度偏移的多项式公式。
// 输入 x = wgsLon - 105.0,y = wgsLat - 35.0(WGS84ToGCJ02 调用时)。
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}
