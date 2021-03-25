package mysql

import (
	"database/sql"
	"errors"
	"yudhiesh/snippetbox/pkg/models"
)

// Define a SnippetModel type which wraps a sql.DB connection pool
type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database
func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	// Statement to insert data to the database
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	// Pass in the placeholder parameters aka the ? in the stmt
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	// Return the id of the inserted record in the snippets table
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// id is an int64 to convert it to a int
	return int(id), nil

}

// Return a single snippet
func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`
	// m.DB.QueryRow returns a pointer to a sql.Row object which holds the
	// result from the database
	row := m.DB.QueryRow(stmt, id)

	// Initialize a pointer to a new zeroed Snippet struct
	s := &models.Snippet{}

	// row.Scan() copies the values from each field to the Snippet struct s,
	// All the values passed are pointers to the place you want to copy the data
	// into, and the number of arguments must be exactly the same as the number
	// of columns returned by your statement
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		// If the query returns no rows then row.Scan() will return a
		// sql.ErrNoRows error.
		// errors.Is() is used to check if the error is a sql.ErrNoRows
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	// If everything is OK then return the Snippet object
	return s, nil
}

// Returns the 10 most recently created snippets
func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return nil, nil
}
