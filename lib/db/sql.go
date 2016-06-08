package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB is a thin wrapper around
type DB struct {
	*sql.DB
}

// Init searches for [decks|users|cards].json to populate table with.
func (db *DB) Init(dir string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Create decks table.
	query := `
    CREATE TABLE IF NOT EXISTS decks(
        Name TEXT PRIMARY KEY,
        InsertedDatetime DATETIME
    );`
	if _, err := tx.Exec(query); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	// Read data from the table.
	b, err := ioutil.ReadFile(dir + "/decks.json")
	if err != nil {
		return err
	}
	decks := make(deckList, 20)
	if err := json.Unmarshal(b, &decks); err != nil {
		return err
	}

	// Store the info in the table
	tx2, err := db.Begin()
	if err != nil {
		return err
	}

	if err := db.Store(decks); err != nil {
		return err
	}

	if err := tx2.Commit(); err != nil {
		return err
	}

	return nil
}

// Open attempts to open an database and will check to make sure it
// can connect to it.  Open doesn't create any tables or populate any
// data into DB (other than what might already exist in filename).
func (db *DB) Open(filename string) error {
	d, err := sql.Open("sqlite3", filename)
	if err != nil {
		return err
	}
	if d == nil {
		return errors.New("db.Open: failed to create db.")
	}
	if err := d.Ping(); err != nil {
		return errors.New("db.Open: failed to connect to db.")
	}

	db.DB = d
	return nil
}

func (db *DB) Store(ls listStorer) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	switch ls := ls.(type) {
	case deckList:
		cmd := `
        INSERT OR REPLACE INTO decks(
            Name, InsertedDatetime
        ) values(?, CURRENT_TIMESTAMP)`
		for _, d := range ls {
			if _, err := tx.Exec(cmd, strings.ToLower(d.Name)); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("db.Store: bad typed (%T) passed in.", ls)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return err
}

func (db *DB) List(l listOp) (listStorer, error) {
	dl := deckList{{Name: "test:deck1"}}
	if strings.HasSuffix(l.query, "*") {
		l.query = strings.TrimRight(l.query, "*")
		l.query += "%"
	}

	// fmt.Println("\n\n\nquery: ", l.query, "\n\n\n")
	switch l.what {
	case "decks":
		cmd := "SELECT Name FROM decks\n"
		cmd += "WHERE Name LIKE\"" + l.query + "\"\n"
		cmd += "ORDER BY Name ASC\n"

		rows, err := db.Query(cmd)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result deckList
		for rows.Next() {
			deck := Deck{}
			err := rows.Scan(&deck.Name)
			if err != nil {
				return nil, err
			}
			result = append(result, deck)
			// fmt.Println("\n\n\n deck: ", deck.Name, "\n\n\n")

		}
		return result, nil
	}

	return dl, nil
}
