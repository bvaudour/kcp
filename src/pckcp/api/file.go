package main

import (
	"io/ioutil"
	"strings"
)

func ReadFile(path string) (lines []string, err error) {
	b, e := ioutil.ReadFile(path)
	if e != nil {
		err = e
		lines = []string{}
	} else {
		lines = strings.Split(string(b), "\n")
	}
	return
}

func WriteFile(path string, lines []string) error {
	b := append([]byte(strings.Join(lines, "\n")), '\n')
	return ioutil.WriteFile(path, b, 0644)
}
