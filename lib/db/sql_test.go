package db

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/askcarter/test"
)

var tests = []struct {
	desc string
	fn   func(t *testing.T, db datasource)
}{
	{"User List/Store",
		func(t *testing.T, db datasource) {
			c := test.Checker(t)

			want := userList{
				{Email: "user1@test.com", Name: "Bill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
				{Email: "user2@test.com", Name: "Jill", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
			}
			err := db.Store(want)
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
		}},

	{"Deck List/Store",
		func(t *testing.T, db datasource) {
			c := test.Checker(t)

			want := deckList{
				{Name: "test1:deck1", Desc: "The meaning of life."},
				{Name: "test1:deck2", Desc: "Essential Camus quotes."},
			}
			err := db.Store(want)
			c.Expect(test.EQ, nil, err)

			got, err := db.List(listOp{what: "decks", query: "test1:deck1"})
			c.Expect(test.EQ, nil, err)
			c.Expect(test.EQ, want[:1], got)

			want = append(want, Deck{Name: "test2:deck2", Desc: "Kayne updates."})
			err = db.Store(want)
			c.Expect(test.EQ, nil, err)

			got, err = db.List(listOp{what: "decks", query: "*"})
			c.Expect(test.EQ, nil, err)
			c.Expect(test.EQ, want, got)
		}},

	{"Card List/Store",
		func(t *testing.T, db datasource) {
			c := test.Checker(t)

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
			err := db.Store(want)
			c.Expect(test.EQ, nil, err)

			got, err := db.List(listOp{what: "cards", query: "user1:*"})
			c.Expect(test.EQ, nil, err)
			checkIgnoreIDs(t, want[:5], got.(cardList))

			card := Card{Owner: "user2:deck1", Front: "chicken", Back: "waffles"}
			want = append(want, card)
			err = db.Store(cardList{card})
			c.Expect(test.EQ, nil, err)

			got, err = db.List(listOp{what: "cards", query: "*"})
			c.Expect(test.EQ, nil, err)
			checkIgnoreIDs(t, want, got.(cardList))
		}},

	{"Init DB from disk",
		func(t *testing.T, db datasource) {
			c := test.Checker(t)

			err := db.Init("./testdata")
			c.Expect(test.EQ, nil, err)

			wantUsers := userList{
				{Email: "ai.ngau@gmail.com", Name: "Ai Ngau", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
				{Email: "askcarter@google.com", Name: "Carter", Password: "$2a$10$KgFhp4HAaBCRAYbFp5XYUOKrbO90yrpUQte4eyafk4Tu6mnZcNWiK"},
			}

			got, err := db.List(listOp{what: "users", query: "*"})
			c.Expect(test.EQ, nil, err)
			c.Expect(test.EQ, wantUsers, got.(userList))

			wantDecks := deckList{
				{Name: "ai.ngau@gmail.com:spanish"},
				{Name: "askcarter@google.com:algebra"},
				{Name: "askcarter@google.com:programming"},
			}

			got, err = db.List(listOp{what: "decks", query: "*"})
			c.Expect(test.EQ, nil, err)
			c.Expect(test.EQ, wantDecks, got.(deckList))

			wantCards := cardList{
				{Owner: "ai.ngau@gmail.com:spanish", Front: "feugo", Back: "pretty"},
				{Owner: "ai.ngau@gmail.com:spanish", Front: "futbol", Back: "soccer"},
				{Owner: "ai.ngau@gmail.com:spanish", Front: "que?", Back: "what?"},
				{Owner: "ai.ngau@gmail.com:spanish", Front: "a donde es?", Back: "where is?"},
				{Owner: "ai.ngau@gmail.com:spanish", Front: "hola", Back: "hello"},
				{Owner: "askcarter@google.com:algebra", Front: "x*0", Back: "0"},
				{Owner: "askcarter@google.com:algebra", Front: "x+0", Back: "x"},
				{Owner: "askcarter@google.com:algebra", Front: "x+x", Back: "2x"},
				{Owner: "askcarter@google.com:programming", Front: "favorite programming language", Back: "Go"},
				{Owner: "askcarter@google.com:programming", Front: "public interface", Back: "API"},
			}

			got, err = db.List(listOp{what: "cards", query: "*"})
			c.Expect(test.EQ, nil, err)
			checkIgnoreIDs(t, wantCards, got.(cardList))
		}},
}

func TestSqlDS(t *testing.T) {
	for _, tt := range tests {
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

		t.Run(tt.desc, func(t *testing.T) {
			tt.fn(t, db)
		})
	}
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
