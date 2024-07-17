package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/microsoft/go-mssqldb"
)

type TokenRecord struct {
	ID     int
	TOKEN  string
	TENANT string
	DOMAIN string
}

const (
	dbServer   = "localhost"
	dbPort     = 1433
	dbUser     = ""
	dbPassword = ""
	database   = ""
)

func getDBConnection() *sql.DB {

	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", dbServer, dbUser, dbPassword, dbPort, database)

	// Open DB connection.
	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		log.Fatal("Error creating database connection: ", err.Error())
	}

	// Verify the connection.
	err = db.Ping()
	if err != nil {
		log.Fatal("Error establishing database connection: ", err.Error())
	}
	log.Println("Connected to SQL Server DB")

	return db
}

func getToken(s *server, token string) (*TokenRecord, error) {

	// Retrieve token from cache.
	s.mu.Lock()
	val, ok := s.tokenCache[token]
	s.mu.Unlock()

	if ok {
		log.Println("Token found in cache")
		return &val, nil
	}

	// Retrieve token from database.
	query := "SELECT ID, TOKEN, TENANT, DOMAIN FROM AGENT_ACCESS_TOKEN WHERE TOKEN=@TOKEN"

	db := getDBConnection()

	// Prepare the query.
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Fatal("Error preparing query: ", err.Error())
	}
	defer stmt.Close()

	// Execute the query.
	rows, err := stmt.Query(sql.Named("TOKEN", token))
	if err != nil {
		log.Fatal("Error executing query: ", err.Error())
	}
	defer rows.Close()

	// Retrieve the results.
	var records []TokenRecord
	for rows.Next() {
		var record TokenRecord
		err := rows.Scan(&record.ID, &record.TOKEN, &record.TENANT, &record.DOMAIN)
		if err != nil {
			log.Fatal("Error scanning row: ", err.Error())
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Error retrieving rows: ", err.Error())
	}
	if len(records) == 1 {
		s.mu.Lock()
		s.tokenCache[token] = records[0]
		s.mu.Unlock()
		return &records[0], nil
	}

	return nil, fmt.Errorf("no record found for token: %s", token)
}
