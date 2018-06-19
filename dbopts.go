/*
Contains database operations for the bot.


Released under MIT license, copyright 2018 Tyler Ramer

*/
package main

import (
	"database/sql"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// Table creation SQL strings
const (
	peopleTable = "CREATE TABLE IF NOT EXISTS people(id SERIAL PRIMARY KEY, name TEXT, karma INTEGER, shame INTEGER);"
	alsoTable   = "CREATE TABLE IF NOT EXISTS isalso(id SERIAL PRIMARY KEY, name TEXT, also TEXT);"
)

// Define a global db connection. We don't need to close the db conn - if there's an error we'll try
// to recreate the db connection, but otherwise we don't intend to trash it
var db *sql.DB

// Connect to the DB and test the connection. Because we're using a global DB connection, and because
// database/sql will retry the connection for us, we should only use this to initialize the db connection
func dbConnect() *sql.DB {
	var err error
	db, err = sql.Open("postgres", conStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Error("Trouble connecting to the database, shutting down")
		log.Fatal(err)
	}
	log.WithField("conStr", conStr).Info("Successfully connected to a postgres DB")

	// go ahead and check tables here
	checkTables()
	return db
}

// confirm all database tables exist and exit if they don't try to create them
func checkTables() {

	var result string
	err := db.QueryRow("SELECT 1 FROM people LIMIT 1").Scan(&result)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Error("Could not select from people table, will try to create it now")
			createPeopleTable()
		} else {
			log.Fatal(err)
		}
	}
	err = db.QueryRow("SELECT 1 from isalso LIMIT 1").Scan(&result)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Error("Could not select from isalso table, will try to create it now")
			createAlsoTable()
		} else {
			log.Fatal(err)
		}
	}

}

// creates the "people" table in database
func createPeopleTable() {
	_, err := db.Exec(peopleTable)
	if err != nil {
		log.Error("Problem creating people table")
		log.Fatal(err)
	}
}

// creates the "isalso" table in database
func createAlsoTable() {
	_, err := db.Exec(alsoTable)
	if err != nil {
		log.Error("Problem creating isalso table")
		log.Fatal(err)
	}
}

func karmaRank() {

}

// Handles karma up/down and shame operations
func karmaMod() {

}

func isAlsoAsk() {

}

func isAlsoAdd() {

}
