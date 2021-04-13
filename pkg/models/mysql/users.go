package mysql

import (
	"database/sql"
	"errors"
	"strings"
	"yudhiesh/snippetbox/pkg/models"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *sql.DB
}

// Insert a user into the users table
func (m *UserModel) Insert(name, email, password string) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	stmt := `INSERT INTO users (name, email, hashed_password, created)
    VALUES(?, ?, ?, UTC_TIMESTAMP())`
	_, err = tx.Exec(stmt, name, email, string(hashedPassword))
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				tx.Rollback()
				return models.ErrDuplicateEmail
			}
		}
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// Authenticate the users email and password
func (m *UserModel) Authenticate(email, password string) (int, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return 0, nil
	}
	// Retrieve the id and hashedPassword from the email
	// If no matching email exists, or the user is not active, we return the
	// ErrInvalidCredentials
	var id int
	var hashedPassword []byte
	stmt := `SELECT id, hashed_password FROM users WHERE email = ? AND active = true`
	row := tx.QueryRow(stmt, email)
	err = row.Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return 0, models.ErrInvalidCredentials
		} else {
			tx.Rollback()
			return 0, err
		}
	}

	// Check whether the hashed password and password match
	// If they do not then return ErrInvalidCredentials
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			tx.Rollback()
			return 0, models.ErrInvalidCredentials
		} else {
			tx.Rollback()
			return 0, nil
		}
	}

	err = tx.Commit()
	return id, err
}

func (m *UserModel) Get(id int) (*models.User, error) {
	u := &models.User{}

	stmt := `SELECT id, name, email, created, active FROM users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created, &u.Active)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	return u, nil
}

func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte
	row := m.DB.QueryRow("SELECT hashed_password FROM users WHERE id = ?", id)
	err := row.Scan(&currentHashedPassword)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return models.ErrInvalidCredentials
		} else {
			return err
		}
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	stmt := "UPDATE users SET hashed_password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
	return err
}
