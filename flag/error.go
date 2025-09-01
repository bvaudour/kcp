package flag

import (
	"codeberg.org/bvaudour/kcp/common"
)

type Error string

func (e Error) Error() string {
	return string(e)
}

func NewError(err string, args ...any) error {
	return Error(common.Tr(err, args...))
}
