package cube

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func TestMain(m *testing.M) {
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
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
		log.Fatalf("Failed to create table: %v", err)
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
		log.Fatalf("Failed to insert data: %v", err)
	}

	os.Exit(m.Run())
}

func TestQueryBuilder(t *testing.T) {
	// Test the QueryBuilder
	var users []struct {
		ID    int
		Name  string
		Email string
		Age   int
		Type  int
	}
	qb := NewQueryBuilder().SelectStruct(&users).From("users").Where("age >= ?", 35)
	t.Log(qb.GetStatement())
	rows, err := qb.Query(db)
	if err != nil {
		if err == sql.ErrNoRows{
			t.Log("No Rows")
		}
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

func TestOrderBy(t *testing.T) {
		// Test the QueryBuilder
		var users []struct {
			ID    int
			Name  string
			Email string
			Age   int
			Type  int
		}
		qb := NewQueryBuilder().SelectStruct(&users).From("users").Where("age >= ?", 35).OrderBy(true, "age")
		t.Log(qb.GetStatement())
		rows, err := qb.Query(db)
		if err != nil {
			if err == sql.ErrNoRows{
				t.Log("No Rows")
			}
			t.Fatalf("Query failed: %v", err)
		}
		err = ScanStruct(rows, &users)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		if len(users) != 3 {
			t.Fatalf("Expected 3 users, got %d", len(users))
		}

		if users[0].Age != 55 {
			t.Fatalf("Expected to get age 35 but got something different")
		}
}
