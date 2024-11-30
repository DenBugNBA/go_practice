package main

import (
	"net/http"
	"strconv"
)

var count int

func countHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.Itoa(count)))
		return

	case http.MethodPost:
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		numberString := r.FormValue("count")
		number, err := strconv.Atoi(numberString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("это не число"))
			return
		}
		count += number

		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/count", countHandler)

	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		panic(err)
	}
}
