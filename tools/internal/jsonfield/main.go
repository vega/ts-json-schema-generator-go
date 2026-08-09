// Command jsonfield prints one top-level string field from JSON on stdin.
// Used by tools/bump-tsgo.sh to avoid a jq dependency.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: jsonfield <field>")
		os.Exit(2)
	}
	var value map[string]any
	if err := json.NewDecoder(os.Stdin).Decode(&value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	field, ok := value[os.Args[1]].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "field %q not found\n", os.Args[1])
		os.Exit(1)
	}
	fmt.Println(field)
}
