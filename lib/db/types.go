package db

import "io"

// User stores information about a user including hashed password,
// an email address (which acts as an unique id), and a display name.
type User struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// Decks belong to a User.  The first part of their name specifies a owner.
// So a the name of a deck called 'math' belonging to 'carter@carter.com' would
// be stored as 'carter@carter.com:math'.  Name's must be unique.
type Deck struct {
	Name string `json:"name"`
}

// A Deck can have many flashcards.  There is no checking that a card is unique.
type Card struct {
	ID    int    `json:"id,omitempty"`
	Owner string `json:"owner"`
	Front string `json:"front"`
	Back  string `json:"back"`
}

type cardList []Card
type deckList []Deck
type userList []User

func (dl deckList) List(ds datasource, l listOp) error {
	return nil
}
func (dl deckList) Store(ds datasource, r io.Reader, s string) error {
	return nil
}
func (ul userList) List(ds datasource, l listOp) error {
	return nil
}
func (ul userList) Store(ds datasource, r io.Reader, s string) error {
	return nil
}
func (cl cardList) List(ds datasource, l listOp) error {
	return nil
}
func (cl cardList) Store(ds datasource, r io.Reader, s string) error {
	return nil
}

type listOp struct {
	what, user, query string
}

// listStorers now how to read from and write to a datasource.
type listStorer interface {
	List(datasource, listOp) error
	Store(datasource, io.Reader, string) error
}

// A datasource is a db for our app to interact with.  I made it an
// an interface so that I could mock out DB calls.
type datasource interface {
	Open(file string) error
	Close() error
	Init(dir string) error

	List(listOp) (listStorer, error)
	Store(ls listStorer) error
}
