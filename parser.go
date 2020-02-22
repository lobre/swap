package main

import (
    "io"
    "io/ioutil"
    "fmt"
)

func parse(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
        return fmt.Errorf("failed to read note content: %w", err)
	}

    _, items := lex(b)
    for item := range items {
        if item.typ == itemEOF {
            break
        }
        fmt.Println(item)
    }
    return nil
}
