package gbt2016

import "github.com/sunsky74/gb32960/api"

// GBT2016Body is embedded in V2016 message structs to signal their protocol
// version. It carries no fields of its own; embedding it only promotes
// Version() so each V2016 body satisfies model.MessageBody without
// duplicating the version constant.
type GBT2016Body struct{}

func (b GBT2016Body) Version() api.GBTVersion { return api.V2016 }
