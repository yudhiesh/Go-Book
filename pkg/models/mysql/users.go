package mysql

import (
	"database/sql"
	"yudhiesh/snippetbox/pkg/models"
)

type UserModel struct {
	DB *sql.DB
}

func (u *UserModel) Insert(name, email, password string) error {
	return nil
}

func (u *UserModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

func (u *UserModel) Get(id int) (*models.User, error) {
	return nil, nil
}
