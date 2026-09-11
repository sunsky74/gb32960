package utils

import "math"

const (
	pi   = 3.1415926535897932384626
	axis = 6378245.0              // Krasovsky ellipsoid semi-major axis
	ee   = 0.00669342162296594323 // 1 - e² (ellipsoid eccentricity squared)
)

// Coordinate holds longitude and latitude.
type Coordinate struct {
	Longitude float64
	Latitude  float64
}

// WGS84ToGCJ02 converts WGS84 coordinates to GCJ02 (Mars coordinate system).
// Coordinates outside China are returned unchanged.
// Algorithm matches Java CoordinateUtils.wgs84ToGcj02() exactly:
// Krasovsky ellipsoid parameters with offset (lon-105, lat-35) + ellipsoid correction.
func WGS84ToGCJ02(wgsLon, wgsLat float64) Coordinate {
	if outOfChina(wgsLat, wgsLon) {
		return Coordinate{Longitude: wgsLon, Latitude: wgsLat}
	}

	// Step 1: compute geographic deltas using offset (lon-105, lat-35)
	dLat := transformLat(wgsLon-105.0, wgsLat-35.0)
	dLon := transformLon(wgsLon-105.0, wgsLat-35.0)

	// Step 2: ellipsoid correction (Krasovsky parameters)
	radLat := wgsLat / 180.0 * math.Pi
	magic := 1 - ee*math.Sin(radLat)*math.Sin(radLat)
	sqrtMagic := math.Sqrt(magic)

	// Step 3: final coordinates with ellipsoid correction
	finalLat := wgsLat + (dLat*180.0)/((axis*(1-ee))/(magic*sqrtMagic)*math.Pi)
	finalLon := wgsLon + (dLon*180.0)/(axis/sqrtMagic*math.Cos(radLat)*math.Pi)

	// Step 4: round to 6 decimal places (matches Java BigDecimal.setScale(6, HALF_UP))
	return Coordinate{
		Longitude: math.Round(finalLon*1_000_000) / 1_000_000,
		Latitude:  math.Round(finalLat*1_000_000) / 1_000_000,
	}
}

func outOfChina(lat, lon float64) bool {
	return lon < 72.004 || lon > 137.8347 || lat < 0.8293 || lat > 55.8271
}

// transformLat — polynomial formula for latitude offset.
// Input x = wgsLon - 105.0, y = wgsLat - 35.0 (as called by WGS84ToGCJ02).
func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*math.Pi) + 40.0*math.Sin(y/3.0*math.Pi)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*math.Pi) + 320.0*math.Sin(y/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}

// transformLon — polynomial formula for longitude offset.
// Input x = wgsLon - 105.0, y = wgsLat - 35.0 (as called by WGS84ToGCJ02).
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y
	ret += (20.0*math.Sin(6.0*x*math.Pi) + 20.0*math.Sin(2.0*x*math.Pi)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*math.Pi) + 40.0*math.Sin(x/3.0*math.Pi)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*math.Pi) + 300.0*math.Sin(x/30.0*math.Pi)) * 2.0 / 3.0
	return ret
}
