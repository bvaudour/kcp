package conf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

//Configuration is a parser of .conf file.
//A configuration file is of the form:
//
//    [section1]
//    value1A = …
//    ;comment
//    #another comment
//    [section2]
//    value2A = …
//    value2B = …
//
//Each value can be accessed by the concatenation
//of the section and its name (for ex.: section1.value1A, section2.value2B, etc.).
type Configuration struct {
	t []string
	m map[string]string
	i map[string]int
}

//New returns an empty configuration
func New() *Configuration {
	return &Configuration{
		m: make(map[string]string),
		i: make(map[string]int),
	}
}

//Keys returns the list of the available values’ keys.
func (c *Configuration) Keys() []string {
	out := make([]string, 0, len(c.i))
	for i := range c.i {
		out = append(out, i)
	}
	sort.Slice(out, func(i, j int) bool {
		return c.i[out[i]] < c.i[out[j]]
	})
	return out
}

//Position returns the index where the key is defined on the file.
//index = line - 1.
func (c *Configuration) Position(key string) int {
	if p, ok := c.i[key]; ok {
		return p
	}
	return -1
}

//Set modifies the given key with the given value.
func (c *Configuration) Set(key string, value string) (ok bool) {
	if _, ok = c.m[key]; ok {
		c.m[key] = value
	}
	return
}

//Contains checks if the given key is defined.
func (c *Configuration) Contains(key string) bool {
	_, ok := c.i[key]
	return ok
}

//Get returns the value of the given keys.
//If the key is not defined, it returns an empty string.
func (c *Configuration) Get(key string) string {
	return c.m[key]
}

//Read parses a .conf files into the Configuration object.
func (c *Configuration) Read(r io.Reader) {
	b := bufio.NewScanner(r)
	var section string
	i := -1
	for b.Scan() {
		i++
		line := b.Text()
		c.t = append(c.t, line)
		line = strings.TrimSpace(line)
		l := len(line)
		// Line is comment or blank line
		if l == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}
		// line is section header
		if line[0] == '[' && line[l-1] == ']' {
			section = line[1 : l-1]
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 && idx < l-1 {
			key, value := strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
			k := section + "." + key
			c.m[k] = value
			c.i[k] = i
		}
	}
}

//Write encodes the Configuration object into the given file.
func (c *Configuration) Write(w io.Writer) error {
	lines := make([]string, len(c.t))
	copy(lines, c.t)
	for k, v := range c.m {
		i := c.Position(k)
		if i < 0 {
			continue
		}
		l := lines[i]
		idx := strings.Index(l, "=")
		v := strings.TrimSpace(v)
		if len(v) == 0 {
			lines[i] = l[:idx]
		} else {
			lines[i] = fmt.Sprintf("%s= %s", lines[:idx], v)
		}
	}
	for _, l := range c.t {
		if _, err := fmt.Fprintln(w, l); err != nil {
			return err
		}
	}
	return nil
}

//Fusion sets all keys both defined in c & c2 with
//the values of c2 into c.
func (c *Configuration) Fusion(c2 *Configuration) {
	for _, k := range c.Keys() {
		if c2.Contains(k) {
			c.Set(k, c2.Get(k))
		}
	}
}

//Parse decodes a .conf file and returns the decoded result.
func Parse(r io.Reader) *Configuration {
	c := New()
	c.Read(r)
	return c
}

//Load parses the file located at the given path and returns
//the decoded result.
func Load(filepath string) (c *Configuration, err error) {
	var f *os.File
	if f, err = os.Open(filepath); err != nil {
		return
	}
	defer f.Close()
	c = Parse(f)
	return
}

//Save encodes the configuration on the file at the given path.
func Save(filepath string, c *Configuration) (err error) {
	var f *os.File
	if f, err = os.Create(filepath); err != nil {
		return
	}
	defer f.Close()
	err = c.Write(f)
	return
}
