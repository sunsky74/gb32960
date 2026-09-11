package realtime

import (
	"reflect"

	"github.com/sunsky74/gb32960/api"
	"github.com/sunsky74/gb32960/modelutil"
)

// LocationV2025Data is the V2025 vehicle location data (TLV 0x05).
// Compared to V2016 LocationData, ADDS northernFlag, eastFlag, coordinateType,
// and convertLongitude/convertLatitude fields.
// CoordinateType maps to Java CoordinateType: 0x01=WGS84, 0x02=GCJ02, 0x03=OTHER.
type LocationV2025Data struct {
	Valid            bool
	NorthernFlag     bool    // true=north latitude, false=south
	EastFlag         bool    // true=east longitude, false=west
	CoordinateType   byte    // CoordinateType enum
	OriginLongitude  float64 // raw longitude × 10^6
	OriginLatitude   float64 // raw latitude × 10^6
	ConvertLongitude float64
	ConvertLatitude  float64
}

func (m *LocationV2025Data) Version() api.GBTVersion { return api.V2025 }
func (m *LocationV2025Data) Bytes() ([]byte, error) {
	return modelutil.DefaultBytes(api.V2025, m, reflect.TypeOf((*LocationV2025Data)(nil)).Elem())
}
