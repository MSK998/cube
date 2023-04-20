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
			Age INTEGER NOT NULL,
			Type INTEGER NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO users ([name], [email], [age], [Type])
		VALUES
			('Alice', 'alice@example.com', 25, NULL),
			('Bob', 'bob@example.com', 30, NULL),
			('Charlie', 'charlie@example.com', 35, NULL),
			('Dave', 'dave@example.com', 40, NULL),
			('Rob', 'Rob@example.com', 55, 123)
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
		Type  int
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
	if len(users) != 4 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}
}
