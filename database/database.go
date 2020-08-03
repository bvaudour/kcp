package database

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

//Database is the decoded structure
//of a json database of packages.
type Database struct {
	LastUpdate  time.Time
	IgnoreRepos []string
	Packages
}

//New returns a new empty database initialized
//by repositories of the organzation to ignore.
func New(ignored ...string) *Database {
	return &Database{
		IgnoreRepos: ignored,
	}
}

//Decode decodes the given file to the database
func (db *Database) Decode(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(db)
}

//Encode encodes the database to json
//and write it to the given file.
func (db *Database) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(db)
}

//Load decodes the file in the given path and
//returns the decoded database.
func Load(fpath string, ignored ...string) (db *Database, err error) {
	var f *os.File
	db = New(ignored...)
	if f, err = os.Open(fpath); err != nil {
		return
	}
	defer f.Close()
	err = db.Decode(f)
	return
}

//Save writes the database into the file on the given path.
func Save(fpath string, db *Database) (err error) {
	var f *os.File
	if f, err = os.Create(fpath); err != nil {
		return
	}
	defer f.Close()
	return db.Encode(f)
}

//Update updates the database from a github organization.
//If optional user and password are given, requests are done
//with authentification in order to have a better rate limit.
func (db *Database) Update(organization string, opt ...string) (counter Counter, err error) {
	r := NewRepository(organization, opt...)
	u := NewUpdater(db, r)
	return u.Update()
}
