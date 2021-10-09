package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"sort"
)

func countWords(r io.Reader) (map[string]int, error) {
	m := make(map[string]int)

	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanWords)
	for scan.Scan() {
		m[scan.Text()]++
	}

	return m, scan.Err()
}

type count struct {
	Word       string
	Occurences int
}

func sortWords(reverse bool, m map[string]int) []count {
	words := []count{}
	for w, n := range m {
		words = append(words, count{Word: w, Occurences: n})
	}

	if reverse {
		sort.Slice(words, func(i, j int) bool { return words[i].Occurences < words[j].Occurences })
	} else {
		sort.Slice(words, func(i, j int) bool { return words[i].Occurences > words[j].Occurences })
	}

	return words
}

func main() {
	var (
		reverseFlag    bool
		jsonFlag       bool
		cpuProfileFlag string
	)

	flag.BoolVar(&reverseFlag, "reverse", false, "reverse sort order")
	flag.BoolVar(&jsonFlag, "json", false, "output to json format")
	flag.StringVar(&cpuProfileFlag, "cpuprofile", "", "create a cpu profile")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "countwords: count and sort words by their number of occurences.")
		fmt.Fprintln(os.Stderr, "usage: countwords [OPTIONS] IN [OUT]")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Reads from file IN or, if - is given, from standard input.")
		fmt.Fprintln(os.Stderr, "Default is to write to standard output, or to file OUT if given.")
	}
	flag.Parse()

	if cpuProfileFlag != "" {
		f, err := os.Create(cpuProfileFlag)
		if err != nil {
			fatalf("can't create cpu profile: %v", err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			fatalf("can't start cpu profile: %v", err)
		}
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() == 0 {
		fatalf("no input file")
	}

	var (
		in  io.Reader = os.Stdin
		out io.Writer = os.Stdout
	)

	if flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fatalf("open input file:", err)
		}
		defer f.Close()
		in = f
	}

	if flag.NArg() == 2 {
		f, err := os.Create(flag.Arg(1))
		if err != nil {
			fatalf("create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	m, err := countWords(in)
	if err != nil {
		fatalf("can't count words: %v", err)
		os.Exit(1)
	}

	counts := sortWords(reverseFlag, m)
	if jsonFlag {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		enc.Encode(counts)
	} else {
		for _, c := range counts {
			fmt.Fprintf(out, "%16d %s\n", c.Occurences, c.Word)
		}
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "countwords: "+format, args...)
	os.Exit(1)
}
