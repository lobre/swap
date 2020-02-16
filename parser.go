package main

func parse(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read note content")
	}

    l, items := lex(b)
    for {
        item := <-items
        if item.typ == itemEOF {
            break
        }
        fmt.Println(item.val)
    }
    return nil
}
