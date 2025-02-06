package factories

import (
	"time"

	"github.com/GaryHY/leviosa/internal/domain/user/models"
)

var birthdate, _ = time.Parse("2006-01-02", "1998-07-12")

func NewBasicUser(overrides map[string]interface{}) *models.User {
	user := &models.User{
		ID:                  "",
		ID:                  "123e4567-e89b-12d3-a456-426614174000",
		Email:               "john.doe@example.com",
		EmailHash:           "john.doe@example.com",
		Password:            "hashedpassword",
		PasswordHash:        "hashedpassword",
		Picture:             "picture",
		EncryptedCreatedAt:  "2025-02-03",
		EncryptedLoggedInAt: "2025-02-03",
		Role:                "basic",
		EncryptedBirthDate:  "1998-07-12",
		LastName:            "DOE",
		FirstName:           "John",
		Gender:              "M",
		Telephone:           "0123456789",
		PostalCode:          "75000",
		City:                "Paris",
		Address1:            "01 Avenue Jean DUPONT",
		Address2:            "",
		GoogleID:            "google_id",
		AppleID:             "apple_id",
	}
	// Apply overrides
	for key, value := range overrides {
		switch key {
		case "ID":
			user.ID = value.(string)
		case "Email":
			user.Email = value.(string)
		case "EmailHash":
			user.EmailHash = value.(string)
		case "Password":
			user.Password = value.(string)
		case "PasswordHash":
			user.PasswordHash = value.(string)
		case "CreatedAt":
			user.CreatedAt = value.(time.Time)
		case "EncryptedCreatedAt":
			user.EncryptedCreatedAt = value.(string)
		case "LoggedInAt":
			user.LoggedInAt = value.(time.Time)
		case "EncryptedLoggedInAt":
			user.EncryptedLoggedInAt = value.(string)
		case "Role":
			user.Role = value.(string)
		case "BirthDate":
			user.BirthDate = value.(time.Time)
		case "EncryptedBirthDate":
			user.EncryptedBirthDate = value.(string)
		case "LastName":
			user.LastName = value.(string)
		case "FirstName":
			user.FirstName = value.(string)
		case "Gender":
			user.Gender = value.(string)
		case "Telephone":
			user.Telephone = value.(string)
		case "PostalCode":
			user.PostalCode = value.(string)
		case "City":
			user.City = value.(string)
		case "Address1":
			user.Address1 = value.(string)
		case "Address2":
			user.Address2 = value.(string)
		case "GoogleID":
			user.GoogleID = value.(string)
		case "AppleID":
			user.AppleID = value.(string)
		}
	}
	return user
}

var (
	Johndoe = &models.User{
		ID:         "1",
		Email:      "john.doe@gmail.com",
		Password:   "$a9rfNhA$N$A78#m",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       models.BASIC.String(),
		BirthDate:  birthdate,
		LastName:   "DOE",
		FirstName:  "John",
		Gender:     "M",
		Telephone:  "0123456789",
	}
	Janedoe = &models.User{
		ID:         "2",
		Email:      "jane.doe@gmail.com",
		Password:   "w4w3f09QF&h)#fwe",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       models.BASIC.String(),
		BirthDate:  birthdate,
		LastName:   "DOE",
		FirstName:  "Jane",
		Gender:     "F",
		Telephone:  "0123456780",
	}
	Jeandoe = &models.User{
		ID:         "1",
		Email:      "jean.doe@gmail.com",
		Password:   "wf0fT^9f2$$_aewa",
		CreatedAt:  time.Now().Add(-time.Hour * 4),
		LoggedInAt: time.Now().Add(-time.Hour * 4),
		Role:       models.BASIC.String(),
		BirthDate:  birthdate,
		LastName:   "DOE",
		FirstName:  "Jean",
		Gender:     "M",
		Telephone:  "0123456781",
	}
)

var Users = map[int]*models.User{
	1: {ID: "1", Email: "john.doe@gmail.com", Password: "$a9rfNhA$N$A78#m", CreatedAt: time.Now().Add(-time.Hour * 4), LoggedInAt: time.Now().Add(-time.Hour * 4), Role: models.BASIC.String(), BirthDate: birthdate, LastName: "DOE", FirstName: "John", Gender: "M", Telephone: "0123456789"},
	2: {ID: "2", Email: "jane.doe@gmail.com", Password: "w4w3f09QF&h)#fwe", CreatedAt: time.Now().Add(-time.Hour * 4), LoggedInAt: time.Now().Add(-time.Hour * 4), Role: models.BASIC.String(), BirthDate: birthdate, LastName: "DOE", FirstName: "Jane", Gender: "F", Telephone: "0123456780"},
	3: {ID: "1", Email: "jean.doe@gmail.com", Password: "wf0fT^9f2$$_aewa", CreatedAt: time.Now().Add(-time.Hour * 4), LoggedInAt: time.Now().Add(-time.Hour * 4), Role: models.BASIC.String(), BirthDate: birthdate, LastName: "DOE", FirstName: "Jean", Gender: "M", Telephone: "0123456781"},
}
