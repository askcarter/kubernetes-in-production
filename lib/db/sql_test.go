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

	want := deckList{{Name: "test:deck1"}}
	var db datasource = &DB{}
	err = db.Open(f.Name())
	c.Expect(test.EQ, nil, err)

	err = db.Init("./testdata")
	c.Expect(test.EQ, nil, err)

	db.Store(want)
	c.Expect(test.EQ, nil, err)

	got, err := db.List(listOp{what: "decks", user: "test", query: "test:deck1"})
	c.Expect(test.EQ, nil, err)
	c.Expect(test.EQ, want, got)

	want = append(want, Deck{Name: "test:deck2"})
	db.Store(want)
	c.Expect(test.EQ, nil, err)

	got, err = db.List(listOp{what: "decks", user: "test", query: "test:*"})
	c.Expect(test.EQ, nil, err)
	c.Expect(test.EQ, want, got)
}
