package main

import (
	"cmp"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

type DatasetUsers struct {
	XMLName xml.Name       `xml:"root"`
	Users   []*DatasetUser `xml:"row"`
}

type DatasetUser struct {
	Id        int    `xml:"id" json:"Id"`
	Name      string `xml:"-" json:"Name"`
	Age       int    `xml:"age" json:"Age"`
	FirstName string `xml:"first_name" json:"-"`
	LastName  string `xml:"last_name" json:"-"`
	About     string `xml:"about" json:"About"`
	Gender    string `xml:"gender" json:"Gender"`
}

type ErrResp struct {
	Error string
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != "ok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	users, err := getUsers()
	if err != nil {
		panic(err)
	}
	prepareUsers(users)
	vals := r.URL.Query()
	users = getFilteredUsers(users, vals.Get("query"))
	err = sortUsers(users, vals.Get("order_field"), vals.Get("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errBts, err := json.Marshal(ErrResp{Error: err.Error()})
		if err != nil {
			slog.Error(err.Error())
		}
		w.Write(errBts)
		return
	}
	users, err = getFilteredByOffsetAndLimit(users, vals.Get("offset"), vals.Get("limit"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	usersBts, err := json.Marshal(users)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	_, err = w.Write(usersBts)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
}

func getUsers() ([]*DatasetUser, error) {
	xmlBytes, err := os.ReadFile("dataset.xml")
	if err != nil {
		return nil, fmt.Errorf("error reading dataset: %w", err)
	}
	var data *DatasetUsers
	err = xml.Unmarshal(xmlBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling dataset: %w", err)
	}
	return data.Users, nil
}

func prepareUsers(users []*DatasetUser) {
	for _, u := range users {
		u.Name = u.FirstName + " " + u.LastName
	}
}

func getFilteredUsers(users []*DatasetUser, q string) []*DatasetUser {
	filteredUsers := make([]*DatasetUser, 0)
	for _, u := range users {
		if strings.Contains(u.Name, q) || strings.Contains(u.About, q) {
			filteredUsers = append(filteredUsers, u)
		}
	}
	return filteredUsers
}

func sortUsers(users []*DatasetUser, orderFld, orderDir string) error {
	if orderDir == "" {
		return nil
	}
	switch orderFld {
	case "Id":
		slices.SortFunc(users, func(a, b *DatasetUser) int {
			if orderDir == "1" {
				return cmp.Compare(b.Id, a.Id)
			}
			return cmp.Compare(a.Id, b.Id)
		})
	case "Age":
		slices.SortFunc(users, func(a, b *DatasetUser) int {
			if orderDir == "1" {
				return cmp.Compare(b.Age, a.Age)
			}
			return cmp.Compare(a.Age, b.Age)
		})
	case "Name", "":
		slices.SortFunc(users, func(a, b *DatasetUser) int {
			if orderDir == "1" {
				return cmp.Compare(b.Name, a.Name)
			}
			return cmp.Compare(a.Name, b.Name)
		})
	default:
		return fmt.Errorf("ErrorBadOrderField")
	}
	return nil
}

func getFilteredByOffsetAndLimit(users []*DatasetUser, offset, limit string) ([]*DatasetUser, error) {
	if offset != "" {
		offNum, err := strconv.Atoi(offset)
		if err != nil {
			return nil, fmt.Errorf("error parsing offset: %w", err)
		}
		if len(users) < offNum {
			return []*DatasetUser{}, nil
		}
		users = users[offNum:]
	}
	if limit != "" {
		limitNum, err := strconv.Atoi(limit)
		if err != nil {
			return nil, fmt.Errorf("error parsing limit: %w", err)
		}
		if len(users) < limitNum {
			return users, nil
		}
		users = users[:limitNum]
	}
	return users, nil
}

func writeErr(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	errBts, err := json.Marshal(ErrResp{Error: err.Error()})
	if err != nil {
		slog.Error(err.Error())
	}
	_, err = w.Write(errBts)
	if err != nil {
		slog.Error(err.Error())
	}
}

func main() {
	http.HandleFunc("/users", SearchServer)

	slog.Info("starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		slog.Error(fmt.Sprintf("error starting server: %v", err.Error()))
	}
}
