// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

	searchquery "github.com/kamichidu/go-gae-search-query"
)

const (
	pkgName = "main"
)

func main() {
	flag.Parse()

	s := strings.Join(flag.Args(), " ")
	v, err := searchquery.Parse(s)
	if err != nil {
		panic(err)
	}
	je := json.NewEncoder(os.Stdout)
	je.SetIndent("", "  ")
	if err := je.Encode(v); err != nil {
		panic(err)
	}
}
