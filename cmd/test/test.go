package main

import (
	"bufio"
	"fmt"
	"strings"
)

var str = `a b c

etusretuiac
etauisretauirn
etuisaetrauisrn
`

func main() {
	r := strings.NewReader(str)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		for _, f := range fields {
			fmt.Printf("'%s'\n", f)
		}
	}
}
