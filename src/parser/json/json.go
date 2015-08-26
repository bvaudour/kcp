//Package json provides utilities to easilly parse a json object.
package json

import (
	"bytes"
	sjson "encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	notABool    = "Not a boolean!"
	notANil     = "Not a nil value!"
	notANumber  = "Not a number!"
	notAString  = "Not a string!"
	notAnArray  = "Not an array!"
	notAnObject = "Not an object!"
	unknownKey  = "Unknown key: %s"
)

//Value represents any value's part of a json.
type Value struct {
	data interface{}
}

//Object represents a map of values of a json.
type Object map[string]interface{}

//Object parses the value into an object.
func (v *Value) Object() (o Object, e error) {
	if m, ok := v.data.(map[string]interface{}); ok {
		o = Object(m)
	} else {
		e = errors.New(notAnObject)
	}
	return
}

//Array parses the value into an array of values.
func (v *Value) Array() (c []*Value, e error) {
	if a, ok := v.data.([]interface{}); ok {
		c = make([]*Value, len(a))
		for i, elt := range a {
			c[i] = &Value{elt}
		}
	} else {
		e = errors.New(notAnArray)
	}
	return
}

//Float64 parses the value into a float64.
func (v *Value) Float64() (c float64, e error) {
	if f, ok := v.data.(float64); ok {
		c = f
	} else if f, ok := v.data.(int64); ok {
		c = float64(f)
	} else {
		e = errors.New(notANumber)
	}
	return
}

//Int64 parses the value into an int64.
func (v *Value) Int64() (c int64, e error) {
	if f, err := v.Float64(); err == nil {
		c = int64(f)
	} else {
		e = err
	}
	return
}

//Bool parses the value into a boolean.
func (v *Value) Bool() (c bool, e error) {
	if b, ok := v.data.(bool); ok {
		c = b
	} else {
		e = errors.New(notABool)
	}
	return
}

//String parses the value into a string.
func (v *Value) String() (c string, e error) {
	if s, ok := v.data.(string); ok {
		c = s
	} else {
		e = errors.New(notAString)
	}
	return
}

//Nil checks if the value is nil.
func (v *Value) Nil() error {
	if v.data == nil {
		return nil
	}
	return errors.New(notANil)
}

//Conversions
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

//ParseReader decodes the json reader into a value.
func ParseReader(r io.Reader) (v *Value, e error) {
	d := sjson.NewDecoder(r)
	v = new(Value)
	e = d.Decode(&v.data)
	return
}

//ParseBytes decodes the json byte slice into a value.
func ParseBytes(b []byte) (*Value, error) { return ParseReader(bytes.NewReader(b)) }

//ObjectReader decodes the json reader into an object.
func ObjectReader(r io.Reader) (Object, error) { return v2o(ParseReader(r)) }

//ObjectBytes decodes the json byte slice into an object.
func ObjectBytes(b []byte) (Object, error) { return v2o(ParseBytes(b)) }

//ArrayObjectReader decodes the json reader into a slice of objects.
func ArrayObjectReader(r io.Reader) ([]Object, error) { return v2a(ParseReader(r)) }

//ArrayObjectBytes decodes the json byte slice into a slice of objects.
func ArrayObjectBytes(b []byte) ([]Object, error) { return v2a(ParseBytes(b)) }

//Marshal encodes the given interface to json.
func Marshal(v interface{}) ([]byte, error) { return sjson.Marshal(v) }

//Marshal encodes the value to json.
func (v *Value) Marshal() ([]byte, error) { return Marshal(v.data) }

//Marshal encodes the object to json.
func (o Object) Marshal() ([]byte, error) { return Marshal(o) }

//Search in childs.
func (o Object) get(k string) (c *Value, e error) {
	if v, ok := o[k]; ok {
		c = &Value{v}
	} else {
		e = fmt.Errorf(unknownKey, k)
	}
	return
}

//Get searches the value recursively with given keys.
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

//GetObject searches the object recursively with given keys.
func (o Object) GetObject(keys ...string) (Object, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Object()
	}
	return nil, e
}

//GetArray searches the slice of values recursively with given keys.
func (o Object) GetArray(keys ...string) ([]*Value, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Array()
	}
	return nil, e
}

//GetFloat64 searches the float recursively with given keys.
func (o Object) GetFloat64(keys ...string) (float64, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Float64()
	}
	return 0, e
}

//GetInt64 searches the int recursively with given keys.
func (o Object) GetInt64(keys ...string) (int64, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Int64()
	}
	return 0, e
}

//GetBool searches the boolean recursively with given keys.
func (o Object) GetBool(keys ...string) (bool, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Bool()
	}
	return false, e
}

//GetString searches the string recursively with given keys.
func (o Object) GetString(keys ...string) (string, error) {
	v, e := o.Get(keys...)
	if e == nil {
		return v.String()
	}
	return "", e
}

//GetString searches the nil value recursively with given keys.
func (o Object) GetNull(keys ...string) error {
	v, e := o.Get(keys...)
	if e == nil {
		return v.Nil()
	}
	return e
}

//Map converts the object to a map of interfaces.
func (o Object) Map() map[string]interface{} { return o }

//String converts the object to a string.
func (o Object) String() string {
	b, e := sjson.Marshal(o)
	if e != nil {
		return e.Error()
	}
	return string(b)
}
