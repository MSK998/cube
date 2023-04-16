package cube

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestQueryBuilder(t *testing.T) {
	// Open a test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE users (
			ID INTEGER PRIMARY KEY,
			Name TEXT NOT NULL,
			Email TEXT NOT NULL UNIQUE,
			Age INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users ([name], [email], [age])
		VALUES
			('Alice', 'alice@example.com', 25),
			('Bob', 'bob@example.com', 30),
			('Charlie', 'charlie@example.com', 35),
			('Dave', 'dave@example.com', 40)
	`)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}

	// Test the QueryBuilder
	var users []struct {
		ID    int
		Name  string
		Email string
		Age   int
	}
	qb := NewQueryBuilder().SelectStruct(&users).From("users").Where("age >= ?", 30)
	rows, err := qb.Query(db)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	err = ScanStruct(rows, &users)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if len(users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}
	if users[0].ID != 2 || users[1].ID != 3 || users[2].ID != 4 {
		t.Fatalf("Unexpected user IDs: %v", users)
	}
	if users[0].Name != "Bob" || users[1].Name != "Charlie" || users[2].Name != "Dave" {
		t.Fatalf("Unexpected user names: %v", users)
	}
	if users[0].Email != "bob@example.com" || users[1].Email != "charlie@example.com" || users[2].Email != "dave@example.com" {
		t.Fatalf("Unexpected user emails: %v", users)
	}
	if users[0].Age != 30 || users[1].Age != 35 || users[2].Age != 40 {
		t.Fatalf("Unexpected user ages: %v", users)
	}
}
