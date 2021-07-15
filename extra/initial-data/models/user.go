package models

import (
	"fmt"
	"time"

	"github.com/isucon/isucon11-qualify/bench/random"
)

type User struct {
	JIAUserID string
	CreatedAt time.Time
}

func NewUser() User {
	return User{random.UserName(), random.Time()}
}

func (u User) Create() error {
	if _, err := db.Exec("INSERT INTO user VALUES (?,?)", u.JIAUserID, u.CreatedAt); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}
