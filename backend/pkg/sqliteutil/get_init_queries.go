package sqliteutil

import (
	"fmt"
	"time"

	"github.com/hengadev/leviosa/internal/domain/user/models"
	"golang.org/x/crypto/bcrypt"
)

const password = "secret1234"

func GetInitQueries() ([]string, error) {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Failed to created hashpassword")
	}
	birthday, _ := time.Parse(models.BirthdayLayout, "1999-08-20")
	queries := []string{
		fmt.Sprintf(
			`INSERT OR IGNORE INTO users 
            (email, password, createdat, loggedinat, role, birthdate, lastname, firstname, gender, telephone)
            VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s');`,
			"admin-livio@outlook.fr",
			string(hashedpassword),
			time.Now(),
			time.Now(),
			"admin",
			birthday,
			"admin lastname",
			"admin firstname",
			"M",
			"0000000000",
		),
	}
	return queries, nil
}
