package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/askcarter/spacerep/lib/db"
	"github.com/gorilla/mux"
)

func main() {
	var (
		httpAddr = flag.String("http", ":80", "HTTP service address.")
		// healthAddr = flag.String("health", ":81", "Health service address.")
		file = flag.String("f", "test", "DB file")
	)
	flag.Parse()

	adb := &appDB{&db.DB{}}
	if err := adb.ds.Open(*file); err != nil {
		panic(err)
	}
	defer adb.ds.Close()

	// Use a buffered error channel so that handlers can
	// keep processing after throwing errors.
	errChan := make(chan error, 10)
	go func() {
		httpServer := new(http.Server)
		httpServer.Addr = *httpAddr

		r := router(adb)
		httpServer.Handler = loggingHandler(r)

		log.Println("Starting server...")
		log.Printf("HTTP service listening on %s", *httpAddr)

		errChan <- httpServer.ListenAndServe()
	}()

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errChan:
			// Log any errors from our server
			log.Fatal(err)
		case s := <-signalChan:
			// ctrl+c is a clean exit
			log.Println(fmt.Sprintf("Captured %v. Exiting...", s))
			os.Exit(0)
		}
	}
}

func router(adb *appDB) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Handle("/init", appHandler(adb.init)).Methods("POST")
	r.Handle("/list", appHandler(adb.list)).Methods("GET")
	r.Handle("/store", appHandler(adb.store)).Methods("POST")
	return r
}

type appDB struct {
	ds db.DataSource
}

func (a *appDB) init(w http.ResponseWriter, r *http.Request) (int, error) {
	if u := r.URL.Query().Get("user"); u != "admin" {
		return http.StatusUnauthorized, errors.New("appdDB.init(): Only admin can init database.")
	}

	if err := a.ds.Init("./testdata"); err != nil {
		return http.StatusInternalServerError, err
	}

	fmt.Fprintf(w, `{"message": "initialized db"}`)

	return http.StatusOK, nil
}

func (a *appDB) list(w http.ResponseWriter, r *http.Request) (int, error) {
	t := r.URL.Query().Get("type")
	q := r.URL.Query().Get("q")
	u := r.URL.Query().Get("user")

	if u == "" || q == "" || t == "" {
		return http.StatusInternalServerError, errors.New("appdDB.list(): Missing expected param.")
	}

	l := db.ListOp{What: t, User: u, Query: q}
	ls, err := a.ds.List(l)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var b []byte
	switch ls.(type) {
	case db.UserList:
		b, err = json.MarshalIndent(ls.(db.UserList), "", "\t")
	case db.DeckList:
		b, err = json.MarshalIndent(ls.(db.DeckList), "", "\t")
	case db.CardList:
		b, err = json.MarshalIndent(ls.(db.CardList), "", "\t")
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Write(b)

	return http.StatusOK, nil
}

func (a *appDB) store(w http.ResponseWriter, r *http.Request) (int, error) {
	t := r.URL.Query().Get("type")
	u := r.URL.Query().Get("user")
	if u == "" {
		return http.StatusInternalServerError, errors.New("appDB.store(): Missing user param.")
	}

	d := json.NewDecoder(r.Body)

	var ls db.ListStorer
	switch t {
	case "users":
		if u != "admin" {
			return http.StatusUnauthorized, errors.New("appDB.store(users): only works for admins.")
		}
		ul := db.UserList{}
		if err := d.Decode(&ul); err != nil {
			return http.StatusInternalServerError, err
		}
		ls = ul
	case "cards":
		ul := db.CardList{}
		if err := d.Decode(&ul); err != nil {
			return http.StatusInternalServerError, err
		}
		ls = ul
	case "decks":
		ul := db.DeckList{}
		if err := d.Decode(&ul); err != nil {
			return http.StatusInternalServerError, err
		}
		ls = ul
	default:
		return http.StatusInternalServerError, errors.New("appDB.store(): Invalid type param.")
	}

	if err := a.ds.Store(ls); err != nil {
		return http.StatusInternalServerError, err
	}

	fmt.Fprintf(w, `{"message": "stored data"}`)

	return http.StatusOK, nil
}

// appHandler server all of this applications web traffic, handling
// error reporting and any setup that might be needed for our requests.
type appHandler func(http.ResponseWriter, *http.Request) (int, error)

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status, err := fn(w, r)
	if err != nil {
		log.Println(err)
	}
	if status >= 400 {
		http.Error(w, http.StatusText(status), status)
	}
}

func loggingHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := "%s - - [%s] \"%s %s %s\" %s\n"
		fmt.Printf(format, r.RemoteAddr, time.Now().Format(time.RFC1123),
			r.Method, r.RequestURI, r.Proto, r.UserAgent())
		h.ServeHTTP(w, r)
	}
}
