package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSearchClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	c := SearchClient{
		AccessToken: "ok",
		URL:         ts.URL,
	}
	cases := []struct {
		name      string
		req       SearchRequest
		usersLen  int
		nextPage  bool
		expectErr bool
	}{
		{
			name:      "get zero users",
			req:       SearchRequest{},
			usersLen:  0,
			nextPage:  true,
			expectErr: false,
		},
		{
			name: "get zero users with offset",
			req: SearchRequest{
				Offset: 35,
			},
			usersLen:  0,
			nextPage:  false,
			expectErr: false,
		},
		{
			name: "get one last user",
			req: SearchRequest{
				Offset: 34,
				Limit:  1,
			},
			usersLen:  1,
			nextPage:  false,
			expectErr: false,
		},
		{
			name: "max limit is 25",
			req: SearchRequest{
				Limit: 30,
			},
			usersLen:  25,
			nextPage:  true,
			expectErr: false,
		},
		{
			name: "query",
			req: SearchRequest{
				Query: "ynn",
				Limit: 5,
			},
			usersLen:  2,
			nextPage:  false,
			expectErr: false,
		},
		{
			name: "query with limit",
			req: SearchRequest{
				Query: "ynn",
				Limit: 1,
			},
			usersLen:  1,
			nextPage:  true,
			expectErr: false,
		},
		{
			name: "query with limit and offset",
			req: SearchRequest{
				Query:  "ynn",
				Limit:  1,
				Offset: 1,
			},
			usersLen:  1,
			nextPage:  false,
			expectErr: false,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			resp, err := c.FindUsers(cs.req)
			if err != nil && !cs.expectErr {
				t.Errorf("got unexpected error: %s", err)
			}
			if err == nil && cs.expectErr {
				t.Error("expected error, got none")
			}
			if err == nil && len(resp.Users) != cs.usersLen {
				t.Errorf("got %d users, expected %d", len(resp.Users), cs.usersLen)
			}
			if err == nil && resp.NextPage != cs.nextPage {
				t.Errorf("got %v NextPage flag, expected %v", resp.NextPage, cs.nextPage)
			}
		})
	}
}

func TestSearchClientErrors(t *testing.T) {
	cases := []struct {
		name    string
		c       *SearchClient
		req     SearchRequest
		handler http.HandlerFunc
		errText string
	}{
		{
			name: "invalid limit",
			c:    new(SearchClient),
			req: SearchRequest{
				Limit: -1,
			},
			errText: "limit must be > 0",
		},
		{
			name: "invalid offset",
			c:    new(SearchClient),
			req: SearchRequest{
				Offset: -1,
			},
			errText: "offset must be > 0",
		},
		{
			name:    "get err from req",
			c:       new(SearchClient),
			handler: nil,
			errText: "unknown error",
		},
		{
			name: "bad access token",
			c: &SearchClient{
				AccessToken: "bad",
			},
			handler: http.HandlerFunc(SearchServer),
			errText: "bad AccessToken",
		},
		{
			name: "err timeout",
			c:    new(SearchClient),
			handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
				time.Sleep(2 * time.Second)
			}),
			errText: "timeout for",
		},
		{
			name: "internal server error",
			c:    new(SearchClient),
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			errText: "SearchServer fatal error",
		},
		{
			name: "bad req: ErrorBadOrderField",
			c: &SearchClient{
				AccessToken: "ok",
			},
			req: SearchRequest{
				OrderField: "bad_fld",
			},
			handler: http.HandlerFunc(SearchServer),
			errText: "OrderField",
		},
		{
			name: "bad req: err resp",
			c: &SearchClient{
				AccessToken: "ok",
			},
			req: SearchRequest{},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("[]"))
			}),
			errText: "cant unpack error json",
		},
		{
			name: "bad req: unknown err",
			c: &SearchClient{
				AccessToken: "ok",
			},
			req: SearchRequest{},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"Error":"err"}`))
			}),
			errText: "unknown bad request error",
		},
		{
			name: "invalid res",
			c: &SearchClient{
				AccessToken: "ok",
			},
			req: SearchRequest{},
			handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte(`["wrong"]`))
			}),
			errText: "cant unpack result json",
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {
			if cs.handler != nil {
				ts := httptest.NewServer(cs.handler)
				cs.c.URL = ts.URL
			}
			_, err := cs.c.FindUsers(cs.req)
			if err == nil {
				t.Error("expected error, got none")
			}
			if err != nil && !strings.Contains(err.Error(), cs.errText) {
				t.Errorf("got error %s, expected %s", err, cs.errText)
			}
		})
	}

}
