package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/askcarter/spacerep/lib/db"
	"github.com/askcarter/test"
)

var tests = []struct {
	desc               string
	method, path, data string
	expect             string
	status             int
}{
	// init Handler Tests.

	{path: "/init?user=admin", method: "POST",
		data:   "",
		expect: "init called",
		status: http.StatusOK,
		desc:   "init handler test.",
	},
	{path: "/init?user=not-admin", method: "POST",
		data:   "",
		expect: "",
		status: http.StatusUnauthorized,
		desc:   "init w/o 'user=admin' errors out.",
	},
	{path: "/init", method: "POST",
		data:   "",
		expect: "",
		status: http.StatusUnauthorized,
		desc:   "init w/o 'user' param errors out.",
	},

	// List handler tests.

	{path: "/list?type=cards&user=carter&q=programming",
		method: "GET",
		data:   "",
		expect: `list called`,
		status: http.StatusOK,
		desc:   "list handler passed query as is",
	},
	{path: "/list?type=unknown",
		method: "GET",
		data:   ``,
		expect: "",
		status: http.StatusInternalServerError,
		desc:   "list with unknown type errors out.",
	},
	{path: "/list?type=decks",
		method: "GET",
		data:   ``,
		expect: "",
		status: http.StatusInternalServerError,
		desc:   "list with no 'q' param will error out.",
	},

	{path: "/list?type=users&q=test",
		method: "GET",
		data:   ``,
		expect: "",
		status: http.StatusInternalServerError,
		desc:   "list with no 'user' param will error out.",
	},

	// Store handler tests.

	{path: "/store?type=<does-not-matter>",
		method: "POST",
		data:   ``,
		expect: "",
		status: http.StatusInternalServerError,
		desc:   "store with no user param will error out.",
	},
	{path: "/store?type=unknown&user=carter",
		method: "POST",
		data:   ``,
		expect: "",
		status: http.StatusInternalServerError,
		desc:   "store with invalid type param will error out.",
	},

	{path: "/store?type=users&user=admin",
		method: "POST",
		data: `[{"email":"email1@gmail.com","name":"One","password":"$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
                {"email":"email2@gmail.com","name":"Two","password":"$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"}]`,
		expect: "email1@gmail.com One\nemail2@gmail.com Two",
		status: http.StatusOK,
		desc:   "store(user) works as intended.",
	},
	{path: "/store?type=users&user=carter",
		method: "POST",
		data: `[{"email":"email1@gmail.com","name":"One","password":"$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
                {"email":"email2@gmail.com","name":"Two","password":"$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"}]`,
		expect: "",
		status: http.StatusUnauthorized,
		desc:   "store(user) only works for the admin",
	},

	{path: "/store?type=decks&user=aingau",
		method: "POST",
		data:   `[{"name":"aingau:geometry", "desc": "Math learnings."},{"name":"aingau:cooking", "desc": "Favorite recipes!"}]`,
		expect: "aingau:geometry Math learnings.\naingau:cooking Favorite recipes!",
		status: http.StatusOK,
		desc:   "store(deck) works as intended",
	},

	{path: "/store?type=cards&user=aingau",
		method: "POST",
		data:   `[{"owner": "aingau:spanish", "front": "adonde", "back": "where"}]`,
		expect: "aingau:spanish adonde where",
		status: http.StatusOK,
		desc:   "store(card) works as expected.",
	},
}

func TestAppDB_Handlers(t *testing.T) {
	for i, tt := range tests {
		c := test.Checker(t, test.Summary(fmt.Sprintf("With test %v: %s", i, tt.desc)))

		adb := &appDB{&mockDB{new(bytes.Buffer)}}
		mr := router(adb)

		var r *http.Request
		if tt.data != "" {
			data := bytes.NewBuffer([]byte(tt.data))
			r = httptest.NewRequest(tt.method, tt.path, data)
		} else {
			r = httptest.NewRequest(tt.method, tt.path, nil)
		}
		w := httptest.NewRecorder()
		mr.ServeHTTP(w, r)

		buf := adb.ds.(*mockDB)
		c.Expect(test.EQ, strings.TrimSpace(tt.expect), strings.TrimSpace(buf.String()))

		if w.Code != tt.status {
			t.Errorf("%s handler returned %v: %v", tt.path, w.Code, w.Body)
		}

		if tt.status == http.StatusOK {
			c.Expect(test.NE, nil, w.Body)
		}
	}
}

type mockDB struct {
	*bytes.Buffer
}

func (m *mockDB) Open(file string) error { return nil }
func (m *mockDB) Close() error           { return nil }
func (m *mockDB) Init(dir string) error {
	fmt.Fprintf(m, "init called\n")
	return nil
}
func (m *mockDB) List(l db.ListOp) (db.ListStorer, error) {
	fmt.Fprintf(m, "list called")
	var ls db.ListStorer
	switch l.What {
	case "users":
		ls = db.UserList{
			{Email: "user1@test.com", Name: "Bill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
			{Email: "user2@test.com", Name: "Jill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
		}
	case "decks":
		ls = db.DeckList{
			{Name: "test1:deck1", Desc: "The meaning of life."},
			{Name: "test1:deck2", Desc: "Essential Camus quotes."},
		}
	case "cards":
		ls = db.CardList{
			{Owner: "user1:deck1", Front: "big", Back: "small"},
			{Owner: "user1:deck1", Front: "tall", Back: "short"},
			{Owner: "user1:deck1", Front: "ugly", Back: "pretty"},
			{Owner: "user1:deck2", Front: "sky", Back: "blue"},
			{Owner: "user1:deck2", Front: "grass", Back: "green"},
			{Owner: "user2:deck1", Front: "peanut butter", Back: "jelly"},
			{Owner: "user2:deck1", Front: "sausage", Back: "egg"},
			{Owner: "user2:deck1", Front: "burger", Back: "fries"},
		}
	default:
		return nil, fmt.Errorf("mockDB.List(): Bad typed passed in (%v).", l.What)
	}
	return ls, nil
}
func (m *mockDB) Store(ls db.ListStorer) error {
	switch ls := ls.(type) {
	case db.UserList:
		for _, u := range ls {
			fmt.Fprintln(m, u.Email, u.Name)
		}
	case db.DeckList:
		for _, d := range ls {
			fmt.Fprintln(m, d.Name, d.Desc)
		}
	case db.CardList:
		for _, c := range ls {
			fmt.Fprintln(m, c.Owner, c.Front, c.Back)
		}
	default:
		return fmt.Errorf("mockDB.List(): Bad typed passed in (%T).", ls)
	}
	return nil
}
