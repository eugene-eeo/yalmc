package main

import "os"
import "fmt"
import "flag"
import "strconv"
import "strings"

func mustInt(strs []string) []int {
	b := make([]int, len(strs))
	for i, s := range strs {
		n, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Unable to convert arg ", i+1, ": ", s, "to a number")
			os.Exit(1)
		}
		b[i] = n
	}
	return b
}

func main() {
	filename := flag.String("filename", "", "path to code")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	inputs := mustInt(flag.Args())
	fp, err := os.Open(*filename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	mailboxes, errors := compile(fp)
	if errors != nil && len(errors) != 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	if *debug {
		row := make([]string, 10)
		for i := 0; i < 10; i++ {
			for j := 0; j < 10; j++ {
				row[j] = fmt.Sprintf("%3d", mailboxes[i*10+j])
			}
			fmt.Fprintln(os.Stderr, strings.Join(row, " | "))
		}
	}
	ctx := newContextFromSlice(mailboxes)
	ctx.input = inputs
	outputs, err := ctx.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, out := range outputs {
		fmt.Println(out)
	}
}
