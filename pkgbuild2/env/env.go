package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

type environ map[string]expand.Variable

func (env *environ) Get(name string) expand.Variable {
	return (*env)[name]
}

func (env *environ) Each(f func(string, expand.Variable) bool) {
	for n, v := range *env {
		if !f(n, v) {
			return
		}
	}
}

func (env *environ) Set(name string, variable expand.Variable) error {
	(*env)[name] = variable

	return nil
}

func NewEnviron() expand.WriteEnviron {
	env := make(environ)
	expand.ListEnviron(os.Environ()...).Each(func(n string, v expand.Variable) bool {
		return env.Set(n, v) == nil
	})

	return &env
}

func Set(env expand.WriteEnviron, name string, variable expand.Variable, add ...bool) (err error) {
	if len(add) > 0 && add[0] {
		cur := env.Get(name)
		if variable, err = Concat(cur, variable); err != nil {
			return err
		}
	}

	return env.Set(name, variable)
}

func setIndex(env expand.WriteEnviron, name string, idx int, variable expand.Variable, add ...bool) (err error) {
	cur := ToIndexed(env.Get(name))
	l := len(cur.List)

	if idx > l {
		lst := make([]string, idx+1)
		copy(lst[:l], cur.List)
		cur.List = lst
	} else if idx < 0 {
		i := idx + l
		if i < 0 {
			return fmt.Errorf("%d: bad array index", idx)
		}
		idx = l
	}

	if len(add) > 0 && add[0] {
		cur.List[idx] += variable.Str
	} else {
		cur.List[idx] = variable.Str
	}

	return env.Set(name, cur)
}

func Assoc(env expand.WriteEnviron, name, index string, variable expand.Variable, add ...bool) (err error) {
	if variable.Kind == expand.Associative || variable.Kind == expand.Indexed {
		return errors.New("Cannot set array in indexed variable")
	}

	cur := env.Get(name)
	if cur.Kind == expand.Associative {
		a := len(add) > 0 && add[0]
		if !variable.IsSet() {
			if !a {
				delete(cur.Map, index)
			}
		} else {
			if a {
				cur.Map[index] += variable.Str
			} else {
				cur.Map[index] = variable.Str
			}
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

func Literal(env expand.Environ, word *syntax.Word) (string, error) {
	return expand.Literal(config(env), word)
}

func Fields(env expand.Environ, words ...*syntax.Word) ([]string, error) {
	return expand.Fields(config(env), words...)
}
