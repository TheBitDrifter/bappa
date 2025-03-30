package table

import (
	"github.com/TheBitDrifter/mask"
)

type buildTag string

const (
	unsafeTag        buildTag = "u"
	schemaEnabledTag buildTag = "se"
)

type config struct {
	AutoElementTypeRegistrationTableCreation bool
	buildTags                                []buildTag
}

var Config config = config{
	AutoElementTypeRegistrationTableCreation: true,
}

// Controlled by presence of build flag "unsafe."
func (c config) Unsafe() bool {
	for _, tag := range c.buildTags {
		if tag == unsafeTag {
			return true
		}
	}
	return false
}

// Controlled by presence of build flag "schema_enabled."
func (c config) SchemaLess() bool {
	for _, tag := range c.buildTags {
		if tag == schemaEnabledTag {
			return false
		}
	}
	return true
}

// Controlled by presence of build flag "mXXX e.g., m256", via mask package.
func (c config) MaxElementCount() int {
	return mask.MaxBits
}
