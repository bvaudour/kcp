package env

import (
	"errors"
	"strings"

	"mvdan.cc/sh/v3/expand"
)

func NewString(value string) expand.Variable {
	return expand.Variable{
		Kind: expand.String,
		Str:  value,
	}
}

func NewIndexed(values ...string) expand.Variable {
	return expand.Variable{
		Kind: expand.Indexed,
		List: values,
	}
}

func ToString(v expand.Variable) expand.Variable {
	if v.Kind == expand.Indexed {
		return NewString(strings.Join(v.List, " "))
	}
	return NewString(v.Str)
}

func ToIndexed(v expand.Variable) expand.Variable {
	if v.Kind == expand.Indexed {
		return v
	}
	return NewIndexed(v.Str)
}

func Concat(v1, v2 expand.Variable) (result expand.Variable, err error) {
	if v1.Kind == expand.Associative || v2.Kind == expand.Associative {
		err = errors.New("Cannot concat associative array")
	} else if v1.Kind == expand.Indexed || v2.Kind == expand.Indexed {
		var lst []string
		if v1.IsSet() {
			lst = ToIndexed(v1).List
		}
		if v2.IsSet() {
			lst = append(lst, ToIndexed(v2).List...)
		}
		result = NewIndexed(lst...)
	} else {
		result = NewString(v1.Str + v2.Str)
	}

	return
}
