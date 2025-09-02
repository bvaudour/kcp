// Package env provides a shell environment that can be used with the mvdan.cc/sh/v3/expand package.
package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"codeberg.org/bvaudour/kcp/pkgbuild/position"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

// environment implements the expand.Environ and expand.WriteEnviron interfaces
// to store and manage shell variables.
type environment map[string]expand.Variable

// New creates a new environment, pre-populated with the host system's
// environment variables using the idiomatic expand.ListEnviron.
func New() expand.WriteEnviron {
	env := make(environment)
	expand.ListEnviron(os.Environ()...).Each(func(name string, vr expand.Variable) bool {
		return env.Set(name, vr) == nil
	})

	return env
}

// Get retrieves a variable from the environment.
// It is part of the expand.Environ interface.
func (env environment) Get(name string) expand.Variable {
	return env[name]
}

// GetDeep acts as env.Get but follows references until first variable.
func GetDeep(env expand.Environ, name string) expand.Variable {
	out := env.Get(name)
	for out.Kind == expand.NameRef {
		out = env.Get(out.Str)
	}

	return out
}

// GetDeepName returns the ultimate reference name.
func GetDeepName(env expand.Environ, name string) string {
	v := env.Get(name)
	for v.Kind == expand.NameRef {
		name = v.Str
		v = env.Get(name)
	}
	return name
}

// Set adds or updates a variable in the environment.
// It is part of the expand.WriteEnviron interface.
// If the variable already exists and is read-only, an error is returned.
func (env environment) Set(name string, vr expand.Variable) error {
	if existing, ok := env[name]; ok && existing.ReadOnly {
		return fmt.Errorf(errReadonlyVariable, name)
	}
	env[name] = vr
	return nil
}

// Each iterates over the variables in the environment.
// It is part of the expand.Environ interface.
func (env environment) Each(f func(name string, vr expand.Variable) bool) {
	for name, vr := range env {
		if !f(name, vr) {
			return
		}
	}
}

func clone(v expand.Variable) expand.Variable {
	out := expand.Variable{
		Set:  v.Set,
		Kind: v.Kind,
		Str:  v.Str,
	}
	out.List = make([]string, len(v.List))
	copy(out.List, v.List)
	if v.Map != nil {
		out.Map = make(map[string]string)
		for k, e := range v.Map {
			out.Map[k] = e
		}
	}
	return out
}

// Concat concatenates two variables based on a set of rules.
// The returned variable's characteristics depend on the input variables.
func Concat(env expand.WriteEnviron, v1, v2 expand.Variable) expand.Variable {
	// If v1 is read-only or v2 is not set, return v1 unchanged.
	if v1.ReadOnly || !v2.IsSet() {
		return v1
	}

	k1, k2 := v1.Kind, v2.Kind
	if k2 == expand.NameRef {
		v2 = GetDeep(env, v2.Str)
		k2 = v2.Kind
	}
	if !v2.IsSet() {
		return v1
	}

	var out expand.Variable
	v1Old, isRef := v1, k1 == expand.NameRef
	if isRef {
		v1 = GetDeep(env, v1.Str)
		k1 = v1.Kind
	}

	switch k1 {
	case expand.String:
		switch k2 {
		case expand.String:
			out = clone(v1)
			out.Str += v2.Str
		case expand.Indexed:
			out = clone(v2)
			out.List = append([]string{v1.Str}, out.List...)
		}
	case expand.Indexed:
		switch k2 {
		case expand.String:
			out = clone(v1)
			out.List = append(out.List, v2.Str)
		case expand.Indexed:
			out = clone(v1)
			out.List = append(out.List, v2.List...)
		}
	case expand.Associative:
		switch k2 {
		case expand.String:
			out = clone(v1)
			out.Map["0"] = out.Map["0"] + v2.Str
		case expand.Indexed:
			out = clone(v1)
			for i := 0; i < len(v2.List); i += 2 {
				k, v := v2.List[i], ""
				if i+1 < len(v2.List) {
					v = v2.List[i+1]
				}
				out.Map[k] = v
			}
		case expand.Associative:
			out = clone(v1)
			for k, v := range v2.Map {
				out.Map[k] = v
			}
		}
	default:
		out = clone(v2)
	}

	out.Local, out.Exported = v1.Local, v1.Exported
	if isRef {
		env.Set(v1Old.Str, out)
		return v1Old
	}

	return out
}

// Set adds or replaces a variable in the given environment.
// If the optional 'add' parameter is true, the new variable is concatenated
// with the existing one using the Concat function.
func Set(env expand.WriteEnviron, name string, variable expand.Variable, add ...bool) error {
	if len(add) > 0 && add[0] {
		variable = Concat(env, env.Get(name), variable)
	}

	return env.Set(name, variable)
}

// ToIndexed converts a variable to an indexed array if it's not already one.
// A string variable becomes a single-element indexed array.
func ToIndexed(env expand.Environ, v expand.Variable) expand.Variable {
	if v.Kind == expand.NameRef {
		v = GetDeep(env, v.Str)
	}
	v = clone(v)
	if v.Kind != expand.Indexed {
		v.Kind, v.List = expand.Indexed, []string{v.Str}
		v.Str, v.Map = "", nil
	}
	return v
}

// ToString converts a variable to a string, emulating bash's behavior.
// If the variable is an indexed array, it returns the first element.
func ToString(env expand.Environ, v expand.Variable) expand.Variable {
	if v.Kind == expand.NameRef {
		v = GetDeep(env, v.Str)
	}
	v = clone(v)
	switch v.Kind {
	case expand.Indexed:
		if len(v.List) > 0 {
			v.Str = v.List[0]
		}
		v.List = v.List[:0]
	case expand.Associative:
		v.Map = nil
	}
	v.Kind = expand.String

	return v
}

func setIndex(env expand.WriteEnviron, name string, idx int, variable expand.Variable, add ...bool) (err error) {
	cur := ToIndexed(env, env.Get(name))
	l := len(cur.List)

	if idx > l {
		lst := make([]string, idx+1)
		copy(lst[:l], cur.List)
		cur.List = lst
	} else if idx < 0 {
		i := idx + l
		if i < 0 {
			return fmt.Errorf(errBadIndex, idx)
		}
		idx = i
	}

	if len(add) > 0 && add[0] {
		cur.List[idx] += variable.Str
	} else {
		cur.List[idx] = variable.Str
	}

	return env.Set(name, cur)
}

// Assoc set the variable name “variable” at index “index”. If add, variable is added to the index variable.
func Assoc(env expand.WriteEnviron, name, index string, variable expand.Variable, add ...bool) (err error) {
	if variable.Kind == expand.Associative || variable.Kind == expand.Indexed {
		return errors.New(errConvertArrayToIndexed)
	}

	cur := GetDeep(env, name)
	name = GetDeepName(env, name)
	if cur.Kind == expand.Associative {
		a := len(add) > 0 && add[0]
		if !variable.IsSet() {
			if !a {
				delete(cur.Map, index)
			}
		} else if a {
			cur.Map[index] += variable.Str
		} else {
			cur.Map[index] = variable.Str
		}
		return env.Set(name, cur)
	}

	var idx int
	if idx, err = strconv.Atoi(index); err != nil {
		return
	}

	return setIndex(env, name, idx, variable, add...)
}

func config(env expand.Environ) *expand.Config {
	return &expand.Config{
		Env: env,
	}
}

// Literal expands a single shell word.
func Literal(env expand.Environ, word *syntax.Word) (string, error) {
	return expand.Literal(config(env), word)
}

// Fields expands a number of words as if they were arguments in a shell command.
func Fields(env expand.Environ, words ...*syntax.Word) ([]string, error) {
	return expand.Fields(config(env), words...)
}

// Keys returns all declared variable’s names.
func Keys(env expand.Environ) (out []string) {
	env.Each(func(name string, _ expand.Variable) bool {
		out = append(out, name)
		return true
	})

	return
}

// ParseAssignation modifies the environment from an assignation.
func ParseAssignation(env expand.WriteEnviron, assign *syntax.Assign) (err error) {
	add, name := assign.Append, assign.Name.Value
	var variable expand.Variable
	var lit string

	if assign.Array != nil {
		variable.Kind, variable.List = expand.Indexed, make([]string, len(assign.Array.Elems))
		for i, elem := range assign.Array.Elems {
			if lit, err = Literal(env, elem.Value); err != nil {
				return
			}
			variable.List[i] = lit
		}
	} else if assign.Value != nil {
		if lit, err = Literal(env, assign.Value); err != nil {
			return
		}
		variable.Kind, variable.Str = expand.String, lit
	} else {
		variable.Kind = expand.Unknown
	}

	if assign.Index == nil {
		err = Set(env, name, variable)
	} else {
		switch assign.Index.(type) {
		case *syntax.BinaryArithm, *syntax.UnaryArithm, *syntax.ParenArithm:
			err = fmt.Errorf(errInvalidSyntax, position.RangeString(assign.Index.Pos(), assign.Index.End()))
		default:
			w := assign.Index.(*syntax.Word)
			index := w.Lit()
			err = Assoc(env, name, index, variable, add)
		}
	}

	return
}
