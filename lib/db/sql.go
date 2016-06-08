package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// DB is a thin wrapper around
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

// func readFromDisk(f string) (listStorer, error) {
// 	b, err := ioutil.ReadFile(f)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var ls listStorer
// 	switch {
// 	case strings.HasSuffix(f, "decks.json"):
// 		decks := deckList{}
// 		if err := json.Unmarshal(b, &decks); err != nil {
// 			return nil, err
// 		}
// 		ls = decks
// 	case strings.HasSuffix(f, "users.json"):
// 		users := userList{}
// 		if err := json.Unmarshal(b, &users); err != nil {
// 			return nil, err
// 		}
// 		ls = users
// 	default:
// 		return nil, errors.New("readFromDisk: bad type passed in: " + f)
// 	}
// 	return ls, nil
// }

// Init searches for [decks|users|cards].json to populate table with.
func (db *DB) Init(dir string) error {
	err := db.createTables()
	if err != nil {
		return err
	}

	// // populate tables
	// tx, err := db.Begin()
	// if err != nil {
	// 	return err
	// }
	// files := []string{"decks.json", "users.json"}
	// for _, f := range files {
	// 	f = filepath.Join(dir, f)
	// 	ul, err := readFromDisk(f)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err := db.Store(ul); err != nil {
	// 		return err
	// 	}
	// }

	// if err := tx.Commit(); err != nil {
	// 	return err
	// }

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
	case userList:
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
	case cardList:
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

func (db *DB) List(l listOp) (listStorer, error) {
	if strings.HasSuffix(l.query, "*") {
		l.query = strings.TrimRight(l.query, "*")
		l.query += "%"
	}

	// fmt.Println("\n\n\nquery: ", l.query, "\n\n\n")
	switch l.what {
	case "decks":
		cmd := "SELECT Name FROM decks\n"
		cmd += "WHERE Name LIKE \"" + l.query + "\"\n"
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

		}
		return result, nil
	case "users":
		cmd := "SELECT Email, Name, Password FROM users\n"
		cmd += "WHERE Email LIKE \"" + l.query + "\"\n"
		cmd += "ORDER BY Email ASC\n"

		rows, err := db.Query(cmd)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result userList
		for rows.Next() {
			user := User{}
			err := rows.Scan(&user.Email, &user.Name, &user.Password)
			if err != nil {
				return nil, err
			}
			result = append(result, user)
		}
		return result, nil
	case "cards":
		cmd := "SELECT ID, Owner, Front, Back FROM cards\n"
		cmd += "WHERE Owner LIKE\"" + l.query + "\"\n"
		cmd += "ORDER BY Owner ASC\n"

		rows, err := db.Query(cmd)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result cardList
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

	return nil, errors.New("db.List(): unknown type passed in: " + l.what)
}
