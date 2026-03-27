package main

import (
	"cmp"
	"fmt"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	return handlePath("", out, path, printFiles)
}

func handlePath(prefix string, out io.Writer, path string, printFiles bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	entries, err := f.ReadDir(-1)
	if err != nil {
		return err
	}
	sortedEntries := getSortedEntries(entries, printFiles)
	for i, e := range sortedEntries {
		isLast := i == len(sortedEntries)-1
		entryStr, err := getEntryString(e, prefix, isLast)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, entryStr)
		if err != nil {
			return err
		}
		if e.IsDir() {
			err = handlePath(appendPrefixByBevel(prefix, isLast), out, fmt.Sprintf("%s/%s", path, e.Name()), printFiles)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getSortedEntries(entries []os.DirEntry, printFiles bool) []os.DirEntry {
	res := make([]os.DirEntry, 0)
	for _, e := range entries {
		if !printFiles && !e.IsDir() {
			continue
		}
		res = append(res, e)
	}
	slices.SortFunc(res, func(i, j os.DirEntry) int {
		return cmp.Compare(i.Name(), j.Name())
	})
	return res
}

func getEntryString(e os.DirEntry, prefix string, isLast bool) (string, error) {
	sb := new(strings.Builder)
	sb.WriteString(prefix)
	if isLast {
		sb.WriteString("└───")
	} else {
		sb.WriteString("├───")
	}
	sb.WriteString(e.Name())
	info, err := e.Info()
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		sb.WriteString(getSizeStr(info))
	}
	return sb.String(), nil
}

func getSizeStr(i fs.FileInfo) string {
	if i.Size() == 0 {
		return " (empty)"
	}
	return fmt.Sprintf(" (%db)", i.Size())
}

func appendPrefixByBevel(prefix string, isLast bool) string {
	if isLast {
		return prefix + "\t"
	}
	return prefix + "│\t"
}
