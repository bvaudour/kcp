package database

import (
	"codeberg.org/bvaudour/kcp/common"
	"strings"
)

// Counter is a counter of updated packages.
type Counter struct {
	Updated int
	Deleted int
	Added   int
}

// String returns the string representation of the counter
func (c Counter) String() string {
	var out []string
	if c.Added > 0 {
		out = append(out, common.Tr(msgAdded, c.Added))
	}
	if c.Deleted > 0 {
		out = append(out, common.Tr(msgDeleted, c.Deleted))
	}
	if c.Updated > 0 {
		out = append(out, common.Tr(msgUpdated, c.Updated))
	}

	return strings.Join(out, "\n")
}
