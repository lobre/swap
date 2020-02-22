package main

import (
    "fmt"
    "flag"
    "os"
)

func main() {
    file := flag.String("file", "", "file to parse")
    flag.Parse()

    if *file == "" {
        fmt.Println("file not provided in flags")
        os.Exit(1)
    }

    f, err := os.Open(*file)
    if err != nil {
        fmt.Println("cannot open file provided")
        os.Exit(1)
    }
    defer f.Close()

    if err := parse(f); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    os.Exit(0)
}
