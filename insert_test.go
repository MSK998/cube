package cube

import (
	"database/sql"
	"testing"
)

func TestInsert(t *testing.T) {

	var users []struct {
		ID    int
		Name  string
		Email string
		Age   int
	}

	// Open a test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a test table
	_, err = db.Exec(`
			CREATE TABLE [users] (
				ID INTEGER PRIMARY KEY,
				Name TEXT NOT NULL,
				Email TEXT NOT NULL UNIQUE,
				Age INTEGER NOT NULL
			)
		`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	qb := NewQueryBuilder().Insert("Name", "Email", "Age").Into("users").Values("hello", "hello1234", 69, "hello", "hello233", 69, "hello", "hello2", 69)
	t.Log(qb.GetStatement())

	_, err = qb.Exec(db)
	if err != nil {
		t.Fatalf(err.Error())
	}

	qb = NewQueryBuilder().SelectStruct(&users).From("users").Where("age = ?", 69)
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
}
