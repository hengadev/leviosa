package sqliteutil

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

const password = "secret"

func GetInitQueries() ([]string, error) {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to created hashpassword")
	}
	birthday, _ := time.Parse(user.BirthdayLayout, "1999-08-20")
	queries := []string{
		fmt.Sprintf(
			`INSERT OR IGNORE INTO users 
            (email, password, createdat, loggedinat, role, birthdate, lastname, firstname, gender, telephone, address, city, postalcard)
            VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s');`,
			"admin@admin.fr",
			hashedpassword,
			time.Now(),
			time.Now(),
			"guest",
			birthday,
			"admin lastname",
			"admin firstname",
			"M",
			"0000000000",
			"admin address",
			"admin city",
			"admin postal card",
		),
	}
	return queries, nil
}

func Init(db *sql.DB, queries ...string) error {
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}
