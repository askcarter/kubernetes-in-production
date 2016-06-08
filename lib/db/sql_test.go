package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/askcarter/test"
)

func TestSqlDS(t *testing.T) {
	c := test.Checker(t)

	// Create temp file for use with this test
	f, err := ioutil.TempFile("", "db_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	var db datasource = &DB{}
	err = db.Open(f.Name())
	c.Expect(test.EQ, nil, err)

	err = db.Init("./testdata")
	c.Expect(test.EQ, nil, err)

	t.Run("decks", func(t *testing.T) {
		want := deckList{{Name: "test:deck1"}}
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err := db.List(listOp{what: "decks", user: "test", query: "test:deck1"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)

		want = append(want, Deck{Name: "test:deck2"})
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err = db.List(listOp{what: "decks", user: "test", query: "test:*"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)
	})

	t.Run("users", func(t *testing.T) {
		want := userList{
			{Email: "user1@test.com", Name: "Bill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
			{Email: "user2@test.com", Name: "Jill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
		}
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err := db.List(listOp{what: "users", user: "user1@test.com", query: "*"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)

		want = append(want, User{Email: "user3@test.com", Name: "John", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"})
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err = db.List(listOp{what: "users", user: "user1@test.com", query: "*", admin: true})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)
	})
}
