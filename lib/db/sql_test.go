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

	t.Run("users", func(t *testing.T) {
		want := userList{
			{Email: "user1@test.com", Name: "Bill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
			{Email: "user2@test.com", Name: "Jill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
		}
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err := db.List(listOp{what: "users", query: "user1@test.com"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want[:1], got)

		want = append(want, User{Email: "user3@test.com", Name: "John", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"})
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err = db.List(listOp{what: "users", query: "*"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)
	})

	t.Run("decks", func(t *testing.T) {
		want := deckList{{Name: "test:deck1"}}
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err := db.List(listOp{what: "decks", query: "test:deck1"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)

		want = append(want, Deck{Name: "test:deck2"})
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err = db.List(listOp{what: "decks", query: "test:*"})
		c.Expect(test.EQ, nil, err)
		c.Expect(test.EQ, want, got)
	})

	t.Run("cards", func(t *testing.T) {
		want := cardList{
			{Owner: "user1:deck1", Front: "big", Back: "small"},
			{Owner: "user1:deck1", Front: "tall", Back: "short"},
			{Owner: "user1:deck1", Front: "ugly", Back: "pretty"},
			{Owner: "user1:deck2", Front: "sky", Back: "blue"},
			{Owner: "user1:deck2", Front: "grass", Back: "green"},
			{Owner: "user2:deck1", Front: "peanut butter", Back: "jelly"},
			{Owner: "user2:deck1", Front: "sausage", Back: "egg"},
			{Owner: "user2:deck1", Front: "burger", Back: "fries"},
		}
		err = db.Store(want)
		c.Expect(test.EQ, nil, err)

		got, err := db.List(listOp{what: "cards", query: "user1:*"})
		c.Expect(test.EQ, nil, err)
		checkIgnoreIDs(t, want[:5], got.(cardList))

		card := Card{Owner: "user2:deck1", Front: "chicken", Back: "waffles"}
		want = append(want, card)
		err = db.Store(cardList{card})
		c.Expect(test.EQ, nil, err)

		got, err = db.List(listOp{what: "cards", query: "user2:*"})
		c.Expect(test.EQ, nil, err)
		checkIgnoreIDs(t, want[5:], got.(cardList))
	})
}

func checkIgnoreIDs(t *testing.T, expected, actual cardList) {
	if len(expected) != len(actual) {
		t.Fatalf("Length mismatch.  \nExpect: %v  \nActual: %v", expected, actual)
	}

	for i, exp := range expected {
		// Ignore ID field.
		act := actual[i]
		if exp.Owner != act.Owner || exp.Front != act.Front || exp.Back != act.Back {
			t.Errorf("Card mismatch.  Expect: %v  Actual: %v", exp, act)
		}
	}
}
