package main

import "os"
import "fmt"
import "flag"
import "strconv"
import "strings"
import "path/filepath"

func mustInt(strs []string) []int {
	b := make([]int, len(strs))
	for i, s := range strs {
		n, err := strconv.Atoi(s)
		if err != nil || n > 999 || n < 0 {
			fmt.Fprintln(os.Stderr, "Unable to convert arg ", i+1, ": ", s, "to a number")
			os.Exit(1)
		}
		b[i] = n
	}
	return b
}

func mustOpen(filepath string) (fp *os.File) {
	fp, err := os.Open(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}

func execFile(path string, inputs []int, debug bool) {
	fp := mustOpen(path)
	defer fp.Close()
	mailboxes, errors := compile(fp)
	if errors != nil && len(errors) != 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	if debug {
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

func isliceEq(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, x := range a {
		if b[i] != x {
			return false
		}
	}
	return true
}

func main() {
	filename := flag.String("filename", "", "path to code")
	workers := flag.Int("workers", 4, "no of workers to use")
	batchMode := flag.Bool("batch", false, "batch process mode")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	if !(*batchMode) {
		inputs := mustInt(flag.Args())
		execFile(*filename, inputs, *debug)
		return
	}
	dirname := filepath.Dir(*filename)
	fp := mustOpen(*filename)
	cases := parseBatch(fp)
	dir := mustOpen(dirname)
	files, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, file := range files {
		path := filepath.Join(dirname, file)
		if path == *filename {
			continue
		}
		code, errs := compile(mustOpen(path))
		// failing to compile a single file is a non-fatal error
		// so just continue trying to compile other files
		if errs != nil {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, e)
			}
			continue
		}
		for _, r := range batch(*workers, code, cases) {
			fmt.Println(
				file,
				r.tcase.name,
				r.tcase.input,
				r.output,
				r.tcase.output,
				isliceEq(r.output, r.tcase.output),
				r.cycles,
			)
		}
	}
}
