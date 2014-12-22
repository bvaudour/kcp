package pjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Value struct {
	data interface{}
}

type Object map[string]interface{}

// Extract value data
func (v *Value) Object() (o Object, e error) {
	switch v.data.(type) {
	case map[string]interface{}:
		o = Object(v.data.(map[string]interface{}))
	default:
		e = errors.New("Not an object!")
	}
	return
}
func (v *Value) Array() (c []*Value, e error) {
	switch v.data.(type) {
	case []interface{}:
		data := v.data.([]interface{})
		c = make([]*Value, len(data))
		for i, d := range data {
			c[i] = &Value{d}
		}
	default:
		e = errors.New("Not an array!")
	}
	return
}
func (v *Value) Number() (c float64, e error) {
	switch v.data.(type) {
	case float64:
		c = v.data.(float64)
	default:
		e = errors.New("Not a number!")
	}
	return
}
func (v *Value) Bool() (c bool, e error) {
	switch v.data.(type) {
	case bool:
		c = v.data.(bool)
	default:
		e = errors.New("Not a bool!")
	}
	return
}
func (v *Value) String() (c string, e error) {
	switch v.data.(type) {
	case string:
		c = v.data.(string)
	default:
		e = errors.New("Not a string!")
	}
	return
}
func (v *Value) Nil() error {
	switch v.data.(type) {
	case nil:
		return nil
	default:
		return errors.New("Not a nil value!")
	}
}

// Conversions
func v2o(v *Value, e error) (Object, error) {
	if e != nil {
		return nil, e
	}
	return v.Object()
}
func v2a(v *Value, e error) ([]Object, error) {
	if e != nil {
		return nil, e
	}
	a, err := v.Array()
	if err != nil {
		return nil, err
	}
	out := make([]Object, len(a))
	for i, elt := range a {
		if o, err := elt.Object(); err != nil {
			return nil, err
		} else {
			out[i] = o
		}
	}
	return out, nil
}

// Parser
func ParseReader(r io.Reader) (v *Value, e error) {
	d := json.NewDecoder(r)
	v = new(Value)
	e = d.Decode(&v.data)
	return
}
func ParseBytes(b []byte) (*Value, error) {
	return ParseReader(bytes.NewReader(b))
}
func ObjectReader(r io.Reader) (Object, error) {
	return v2o(ParseReader(r))
}
func ObjectBytes(b []byte) (Object, error) {
	return v2o(ParseBytes(b))
}
func ArrayObjectReader(r io.Reader) ([]Object, error) {
	return v2a(ParseReader(r))
}
func ArrayObjectBytes(b []byte) ([]Object, error) {
	return v2a(ParseBytes(b))
}

// Marshaller
func Marshal(v interface{}) ([]byte, error) { return json.Marshal(v) }
func (v *Value) Marshal() ([]byte, error)   { return Marshal(v.data) }
func (o Object) Marshal() ([]byte, error)   { return Marshal(o) }

// Search in childs
func (o Object) get(k string) (c *Value, e error) {
	if v, ok := o[k]; ok {
		c = &Value{v}
	} else {
		e = errors.New(fmt.Sprintf("Unknown key: %s", k))
	}
	return
}
func (o Object) Get(keys ...string) (*Value, error) {
	oc := o
	for _, k := range keys[:len(keys)-1] {
		v, e := oc.get(k)
		if e != nil {
			return nil, e
		}
		oc, e = v.Object()
		if e != nil {
			return nil, e
		}
	}
	return oc.get(keys[len(keys)-1])
}

// Search in childs - typed
func (o Object) GetObject(keys ...string) (Object, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Object()
	}
	return nil, e
}
func (o Object) GetArray(keys ...string) ([]*Value, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Array()
	}
	return nil, e
}
func (o Object) GetNumber(keys ...string) (float64, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Number()
	}
	return 0, e
}
func (o Object) GetBool(keys ...string) (bool, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Bool()
	}
	return false, e
}
func (o Object) GetString(keys ...string) (string, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.String()
	}
	return "", e
}
func (o Object) GetNull(keys ...string) error {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Nil()
	}
	return e
}

// Map representation
func (o Object) Map() map[string]interface{} { return o }

// String representation
func (o Object) String() string {
	b, e := json.Marshal(o)
	if e != nil {
		return e.Error()
	}
	return string(b)
}
