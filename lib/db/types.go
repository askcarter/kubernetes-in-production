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
// be stored as 'carter@carter.com:math'.  'Name' must be unique.
type Deck struct {
	Name string `json:"name"`
	Desc string `json:"desc,omitempty"`
}

// A Deck can have many flashcards.  There is no checking that a card is unique.
type Card struct {
	ID    int    `json:"id,omitempty"`
	Owner string `json:"owner"`
	Front string `json:"front"`
	Back  string `json:"back"`
}

type CardList []Card
type DeckList []Deck
type UserList []User

func (dl DeckList) List(ds DataSource, l ListOp) error {
	return nil
}
func (dl DeckList) Store(ds DataSource, r io.Reader, s string) error {
	return nil
}
func (ul UserList) List(ds DataSource, l ListOp) error {
	return nil
}
func (ul UserList) Store(ds DataSource, r io.Reader, s string) error {
	return nil
}
func (cl CardList) List(ds DataSource, l ListOp) error {
	return nil
}
func (cl CardList) Store(ds DataSource, r io.Reader, s string) error {
	return nil
}

type ListOp struct {
	What, User, Query string
}

// ListStorers now how to read from and write to a DataSource.
type ListStorer interface {
	List(DataSource, ListOp) error
	Store(DataSource, io.Reader, string) error
}

// A DataSource is a db for our app to interact with.  I made it an
// an interface so that I could mock out DB calls.
type DataSource interface {
	Open(file string) error
	Close() error
	Init(dir string) error

	List(ListOp) (ListStorer, error)
	Store(ls ListStorer) error
}
