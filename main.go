package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
)

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func sortWords(reverse bool, m map[string]int) []WordCount {
	words := []WordCount{}
	for w, n := range m {
		words = append(words, WordCount{Word: w, Count: n})
	}

	if reverse {
		sort.Slice(words, func(i, j int) bool {
			return words[i].Count < words[j].Count
		})
	} else {
		sort.Slice(words, func(i, j int) bool {
			return words[i].Count > words[j].Count
		})
	}
	return words
}

func countWords(r io.Reader) (map[string]int, error) {
	m := make(map[string]int)

	scan := bufio.NewScanner(r)
	scan.Split(bufio.ScanWords)
	for scan.Scan() {
		m[scan.Text()]++
	}

	return m, scan.Err()
}

func main() {
	var (
		reverseFlag bool
		jsonFlag    bool
	)

	flag.BoolVar(&reverseFlag, "reverse", false, "reverse sort order")
	flag.BoolVar(&jsonFlag, "json", false, "output to json format")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "countwords: count and sort words by their number of occurences.")
		fmt.Fprintln(os.Stderr, "usage: countwords [OPTIONS] [IN|-] [OUT]")
		fmt.Fprintln(os.Stderr, "options:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Reads from file IN or, if - is given, from standard input.")
		fmt.Fprintln(os.Stderr, "Default is to write to standard output, or to file OUT if given.")
	}
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("no input file")
	}

	var (
		// Create and set defaults for IN and OUT.
		in  io.Reader = os.Stdin
		out io.Writer = os.Stdout
	)

	if flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("open input file: %s", err)
		}
		defer f.Close()
		in = f
	}

	if flag.NArg() == 2 {
		f, err := os.Create(flag.Arg(1))
		if err != nil {
			log.Fatalf("create output file: %v", err)
		}
		defer f.Close()
		out = f
	}

	m, err := countWords(in)
	if err != nil {
		log.Fatalf("can't count words: %v", err)
	}

	counts := sortWords(reverseFlag, m)
	if jsonFlag {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(counts); err != nil {
			log.Fatalf("json encoding failed: %v", err)
		}
	} else {
		for _, c := range counts {
			fmt.Fprintf(out, "%16d %s\n", c.Count, c.Word)
		}
	}
}
