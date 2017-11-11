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

func printMailboxes(vm *context) {
	row := make([]string, 10)
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			row[j] = fmt.Sprintf("%3d", vm.mem[i*10+j])
		}
		fmt.Fprintln(os.Stderr, strings.Join(row, " | "))
	}
}

func execFile(path string, inputs []int, debug bool) {
	fp := mustOpen(path)
	defer fp.Close()
	mailboxes, _, errors := compile(fp)
	if errors != nil && len(errors) != 0 {
		for _, err := range errors {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
	ctx := newContextFromSlice(mailboxes)
	if debug {
		printMailboxes(ctx)
	}
	ctx.input = inputs
	outputs, err := ctx.run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, out := range outputs {
		fmt.Println(out)
	}
	if debug {
		printMailboxes(ctx)
	}
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
	fmt.Fprintln(os.Stderr, "Reading batch file:", *filename)
	dirname := filepath.Dir(*filename)
	fp := mustOpen(*filename)
	cases, errors := parseBatch(fp)
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintln(os.Stderr, " ", e)
		}
	}
	dir := mustOpen(dirname)
	files, err := dir.Readdirnames(-1)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	table := newTable()
	for _, file := range files {
		path := filepath.Join(dirname, file)
		if path == *filename {
			continue
		}
		code, used, errs := compile(mustOpen(path))
		fmt.Fprintln(os.Stderr, "  Compiling:", file)
		// failing to compile a single file is a non-fatal error
		// so just continue trying to compile other files
		if errs != nil {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "   ", e)
			}
			continue
		}
		table.addRow(path, used, batch(*workers, code, cases))
	}
	err = table.write(os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
