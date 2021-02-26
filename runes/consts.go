package runes

import (
	"github.com/bvaudour/kcp/common"
)

type Error string

const (
	ErrInvalidToken Error = "Invalid token"
	ErrUnendedToken Error = "Unended token"
)

func (e Error) Error() string {
	return common.Tr(string(e))
}
