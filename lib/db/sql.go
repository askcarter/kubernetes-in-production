package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB is a thin wrapper around sql.DB that know hows to operate on
// db types.
type DB struct {
	*sql.DB
}

func (db *DB) createTables() error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Create decks table.
	queries := []string{
		`CREATE TABLE IF NOT EXISTS decks(
            Name TEXT PRIMARY KEY,
            Desc TEXT,
            InsertedDatetime DATETIME
        );`,
		`CREATE TABLE IF NOT EXISTS users(
            Email TEXT PRIMARY KEY,
            Name TEXT,
            Password TEXT,
            InsertedDatetime DATETIME
        );`,
		`CREATE TABLE IF NOT EXISTS cards(
            ID INTEGER PRIMARY KEY,
            Front TEXT,
            Back  TEXT,
            Owner TEXT,
            InsertedDatetime DATETIME
        );`,
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func readFromDisk(f string) (ListStorer, error) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	var ls ListStorer
	switch {
	case strings.HasSuffix(f, "users.json"):
		users := UserList{}
		if err := json.Unmarshal(b, &users); err != nil {
			return nil, err
		}
		ls = users
	case strings.HasSuffix(f, "decks.json"):
		decks := DeckList{}
		if err := json.Unmarshal(b, &decks); err != nil {
			return nil, err
		}
		ls = decks
	case strings.HasSuffix(f, "cards.json"):
		cards := CardList{}
		if err := json.Unmarshal(b, &cards); err != nil {
			return nil, err
		}
		ls = cards
	default:
		return nil, errors.New("readFromDisk: bad type passed in: " + f)
	}
	return ls, nil
}

// Init reads [decks|users|cards].json files from the files the given
// directory, and then stores them in db.
func (db *DB) Init(dir string) error {
	// populate tables
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	files := []string{"decks.json", "users.json", "cards.json"}
	for _, f := range files {
		f = filepath.Join(dir, f)
		ul, err := readFromDisk(f)
		if err != nil {
			return err
		}
		if err := db.Store(ul); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Open attempts to open an database and will check to make sure it
// can connect to it.  After a DB is Open'd it is ready to Store/List data
// from.
//
// Open doesn't populate any data into DB (other than what might already exist
// in filename).
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

	err = db.createTables()
	if err != nil {
		return err
	}

	return nil
}

// Store inserts the elements of ls into db.
func (db *DB) Store(ls ListStorer) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	switch ls := ls.(type) {
	case DeckList:
		cmd := `
        INSERT OR REPLACE INTO decks(
            Name, Desc, InsertedDatetime
        ) values(?, ?, CURRENT_TIMESTAMP)`
		for _, d := range ls {
			if _, err := tx.Exec(cmd, strings.ToLower(d.Name), d.Desc); err != nil {
				return err
			}
		}
	case UserList:
		cmd := `
        INSERT OR REPLACE INTO users(
            Email, Name, Password, InsertedDatetime
        ) values(?, ?, ?, CURRENT_TIMESTAMP)`
		for _, u := range ls {
			e := strings.ToLower(u.Email)
			if _, err := tx.Exec(cmd, e, u.Name, u.Password); err != nil {
				return err
			}
		}
	case CardList:
		cmd := `
        INSERT OR REPLACE INTO cards(
            ID, Front, Back, Owner, InsertedDatetime
        ) values(NULL, ?, ?, ?, CURRENT_TIMESTAMP)`
		for _, c := range ls {
			if _, err := tx.Exec(cmd, c.Front, c.Back, c.Owner); err != nil {
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

// List retrieves ListStorers from the db as specified by a ListOp.
func (db *DB) List(l ListOp) (ListStorer, error) {
	if strings.HasSuffix(l.Query, "*") {
		l.Query = strings.TrimRight(l.Query, "*")
		l.Query += "%"
	}

	switch l.What {
	case "users":
		cmd := `SELECT Email, Name, Password FROM users
                WHERE Email LIKE ?
                ORDER BY Email ASC`

		rows, err := db.Query(cmd, l.Query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result UserList
		for rows.Next() {
			user := User{}
			err := rows.Scan(&user.Email, &user.Name, &user.Password)
			if err != nil {
				return nil, err
			}
			result = append(result, user)
		}
		return result, nil
	case "decks":
		cmd := `SELECT Name, Desc FROM decks
		        WHERE Name LIKE ?
		        ORDER BY Name ASC`

		rows, err := db.Query(cmd, l.Query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result DeckList
		for rows.Next() {
			deck := Deck{}
			err := rows.Scan(&deck.Name, &deck.Desc)
			if err != nil {
				return nil, err
			}
			result = append(result, deck)
		}
		return result, nil
	case "cards":
		cmd := `SELECT ID, Owner, Front, Back FROM cards
		        WHERE Owner LIKE ?
		        ORDER BY Owner ASC`

		rows, err := db.Query(cmd, l.Query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result CardList
		for rows.Next() {
			card := Card{}
			err := rows.Scan(&card.ID, &card.Owner, &card.Front, &card.Back)
			if err != nil {
				return nil, err
			}
			result = append(result, card)
		}
		return result, nil
	}

	return nil, errors.New("db.List(): unknown type passed in: " + l.What)
}
