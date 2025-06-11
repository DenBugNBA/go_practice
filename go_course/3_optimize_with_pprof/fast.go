package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mailru/easyjson"
	"go_practice/go_course/3_optimize_with_pprof/structs"
)

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(out)
	defer w.Flush()
	fmt.Fprint(w, "found users:\n")

	seenBrowsers := make(map[string]struct{})
	var hasAndroid, hasMSIE bool
	var email string
	androidBrowser := "Android"
	msieBrowser := "MSIE"
	atSign := "@"
	atReplaced := " [at] "
	s := bufio.NewScanner(file)
	user := &structs.User{}

	i := 0
	for s.Scan() {
		err = easyjson.Unmarshal(s.Bytes(), user)
		if err != nil {
			panic(err)
		}
		for _, browser := range user.Browsers {
			if strings.Contains(browser, androidBrowser) {
				hasAndroid = true
				seenBrowsers[browser] = struct{}{}
			}
			if strings.Contains(browser, msieBrowser) {
				hasMSIE = true
				seenBrowsers[browser] = struct{}{}
			}
		}
		if hasAndroid && hasMSIE {
			email = strings.ReplaceAll(user.Email, atSign, atReplaced)
			fmt.Fprint(w, fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email))
		}
		hasAndroid = false
		hasMSIE = false
		i++
	}

	fmt.Fprintln(w, "\nTotal unique browsers", len(seenBrowsers))
}
