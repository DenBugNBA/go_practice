package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
)

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
	case "/user/profile":
		srv.wrapperProfile(w, r)
	case "/user/create":
		srv.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}
}

func (srv *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	rawParamLogin := r.FormValue("login")
	paramLogin := rawParamLogin
	if paramLogin == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"login must not be empty\"}"))
		return
	}

	methodParams := ProfileParams{
		Login: paramLogin,
	}
	res, err := srv.Profile(r.Context(), methodParams)
	if err != nil {
		var apiErr ApiError
		if errors.As(err, &apiErr) {
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		}
		return
	}
	resp := make(map[string]any, 2)
	resp["response"] = res
	resp["error"] = ""
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (srv *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\": \"bad method\"}"))
		return
	}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\": \"unauthorized\"}"))
		return
	}
	rawParamLogin := r.FormValue("login")
	paramLogin := rawParamLogin
	if paramLogin == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"login must not be empty\"}"))
		return
	}
	if len(paramLogin) < 10  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"login len must be >= 10\"}"))
		return
	}

	rawParamName := r.FormValue("full_name")
	paramName := rawParamName

	rawParamStatus := r.FormValue("status")
	paramStatus := rawParamStatus
	if paramStatus == "" {
		paramStatus = "user"
	}
	supportedValues := []string{"user", "moderator", "admin"}
	if !slices.Contains(supportedValues, paramStatus) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"status must be one of [user, moderator, admin]\"}"))
		return
	}

	rawParamAge := r.FormValue("age")
	paramAge, err := strconv.Atoi(rawParamAge)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"age must be int\"}"))
		return
	}
	if paramAge < 0  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"age must be >= 0\"}"))
		return
	}
	if paramAge > 128  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"age must be <= 128\"}"))
		return
	}

	methodParams := CreateParams{
		Login: paramLogin,
		Name: paramName,
		Status: paramStatus,
		Age: paramAge,
	}
	res, err := srv.Create(r.Context(), methodParams)
	if err != nil {
		var apiErr ApiError
		if errors.As(err, &apiErr) {
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		}
		return
	}
	resp := make(map[string]any, 2)
	resp["response"] = res
	resp["error"] = ""
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
	case "/user/create":
		srv.wrapperCreate(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("{\"error\": \"unknown method\"}"))
	}
}

func (srv *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("{\"error\": \"bad method\"}"))
		return
	}
	if r.Header.Get("X-Auth") != "100500" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("{\"error\": \"unauthorized\"}"))
		return
	}
	rawParamUsername := r.FormValue("username")
	paramUsername := rawParamUsername
	if paramUsername == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"username must not be empty\"}"))
		return
	}
	if len(paramUsername) < 3  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"username len must be >= 3\"}"))
		return
	}

	rawParamName := r.FormValue("account_name")
	paramName := rawParamName

	rawParamClass := r.FormValue("class")
	paramClass := rawParamClass
	if paramClass == "" {
		paramClass = "warrior"
	}
	supportedValues := []string{"warrior", "sorcerer", "rouge"}
	if !slices.Contains(supportedValues, paramClass) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"class must be one of [warrior, sorcerer, rouge]\"}"))
		return
	}

	rawParamLevel := r.FormValue("level")
	paramLevel, err := strconv.Atoi(rawParamLevel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"level must be int\"}"))
		return
	}
	if paramLevel < 1  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"level must be >= 1\"}"))
		return
	}
	if paramLevel > 50  {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"error\": \"level must be <= 50\"}"))
		return
	}

	methodParams := OtherCreateParams{
		Username: paramUsername,
		Name: paramName,
		Class: paramClass,
		Level: paramLevel,
	}
	res, err := srv.Create(r.Context(), methodParams)
	if err != nil {
		var apiErr ApiError
		if errors.As(err, &apiErr) {
			w.WriteHeader(apiErr.HTTPStatus)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", err.Error())))
		}
		return
	}
	resp := make(map[string]any, 2)
	resp["response"] = res
	resp["error"] = ""
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
