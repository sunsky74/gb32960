// Package gbt2025 defines GB/T 32960.3-2025 protocol message body structs.
package gbt2025

import "github.com/sunsky74/gb32960/api"

// GBT2025Body is embedded in V2025 message structs to signal their protocol version.
type GBT2025Body struct{}

// Version returns V2025.
func (b GBT2025Body) Version() api.GBTVersion { return api.V2025 }
