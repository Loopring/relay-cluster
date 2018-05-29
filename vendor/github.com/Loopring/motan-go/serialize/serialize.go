package serialize

import (
	motan "github.com/Loopring/motan-go/core"
)

const (
	Simple = "simple"
	Gob = "gob"
)

func RegistDefaultSerializations(extFactory motan.ExtentionFactory) {
	extFactory.RegistryExtSerialization(Simple, 6, func() motan.Serialization {
		return &SimpleSerialization{}
	})
	extFactory.RegistryExtSerialization(Gob, 8, func() motan.Serialization {
		return &GobSerialization{}
	})
}
