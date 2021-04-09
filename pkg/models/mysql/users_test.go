package mysql

import (
	"reflect"
	"testing"
	"time"
	"yudhiesh/snippetbox/pkg/models"
)

func TestUserModeGet(t *testing.T) {
	// Skip the test if the `-short` flag is provided
	if testing.Short() {
		t.Skip("mysql: skipping integration test")
	}
	// Set up a suite of table-driven tests and expected results.
	tests := []struct {
		name      string
		userID    int
		wantUser  *models.User
		wantError error
	}{
		{
			name:   "Valid ID",
			userID: 1,
			wantUser: &models.User{
				ID:      1,
				Name:    "Alice Jones",
				Email:   "alice@example.com",
				Created: time.Date(2018, 12, 23, 17, 25, 22, 0, time.UTC),
				Active:  true,
			},
			wantError: nil,
		},
		{
			name:      "Zero ID",
			userID:    0,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
		{
			name:      "Non-existent ID",
			userID:    2,
			wantUser:  nil,
			wantError: models.ErrNoRecord,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the new test database
			db, teardown := newTestDB(t)
			// run the teardown function at the end
			defer teardown()

			m := UserModel{db}

			user, err := m.Get(tt.userID)
			if err != tt.wantError {
				t.Errorf("want %v; got %s", tt.wantError, err)
			}
			// Check for equality between arbitrarily complex custom types
			if !reflect.DeepEqual(user, tt.wantUser) {
				t.Errorf("want %v; got %v", tt.wantUser, user)
			}

		})
	}
}
