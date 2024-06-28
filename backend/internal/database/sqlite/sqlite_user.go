package sqlite

import (
	"log"

	"github.com/GaryHY/event-reservation-app/internal/types"
	"golang.org/x/crypto/bcrypt"
)

func (s *Store) IsAdmin(session_id string) bool {
	var role string
	query := `
        SELECT role FROM users WHERE email = 
        (SELECT email FROM sessions WHERE id = ?);
    `
	err := s.DB.QueryRow(query).Scan(&role)
	if err != nil {
		log.Fatalf("Failed to find if the user refered to the sessions id %q is an admin user", session_id)
	}
	return types.ConvertToRole(role) == types.ADMIN
}

// Function that returns true if user in database already
func (s *Store) CheckUser(email string) bool {
	var count int
	s.DB.QueryRow("SELECT 1 FROM users WHERE email=? ;", email).Scan(&count)
	return count == 1
}

func (s *Store) CheckUserById(user_id string) bool {
	var count int
	s.DB.QueryRow("SELECT 1 FROM users WHERE id=? ;", user_id).Scan(&count)
	return count == 1
}

func (s *Store) DeleteUser(user_id string) error {
	_, err := s.DB.Exec("DELETE FROM users WHERE id = ?", user_id)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateUser(user *types.UserStored) error {
	query := `
        UPDATE users SET 
		email=?,
		password=?,
		role=?,
		lastName=?,
		firstName=?,
		gender=?,
		birthDate=?,
		telephone=?,
		address=?,
		city=?,
		postalCard=?,
        WHERE id=?;
    `
	_, err := s.DB.Exec(
		query,
		user.Email,
		user.Password,
		user.Role,
		user.LastName,
		user.FirstName,
		user.Gender,
		user.BirthDate,
		user.Telephone,
		user.Address,
		user.City,
		user.PostalCard,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetHashPassword(user *types.User) (hashpassword string) {
	s.DB.QueryRow("SELECT hashpassword from users where email = ?;", user.Email).Scan(&hashpassword)
	return
}

func (s *Store) GetUserId(user_email string) (id string) {
	s.DB.QueryRow("SELECT id from users where email = ?;", user_email).Scan(&id)
	return
}

// TODO: Change that function once all the field are fine !
func (s *Store) CreateUser(newUser *types.UserStored) error {
	hashpassword := hashPassword(newUser.Password)
	_, err := s.DB.Exec("INSERT INTO users (id, email, hashpassword, role, lastname, firstname, gender, birthdate, telephone, address, city, postalcard) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);", newUser.Id, newUser.Email, hashpassword, newUser.Role, newUser.LastName, newUser.FirstName, newUser.Gender, newUser.BirthDate, newUser.Telephone, newUser.Address, newUser.City, newUser.PostalCard)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetUserIdBySessionId(session_id string) (id string) {
	s.DB.QueryRow("SELECT userid from sessions where id = ?;", session_id).Scan(&id)
	return
}

func (s *Store) GetAllUsers() []*types.UserStored {
	users := make([]*types.UserStored, 0)
	rows, err := s.DB.Query("SELECT * FROM users;")
	if err != nil {
		log.Fatal("Cannot get events rows - ", err)
	}
	defer rows.Close()

	for rows.Next() {
		user := &types.UserStored{}
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.LastName,
			&user.FirstName,
			&user.Gender,
			&user.BirthDate,
			&user.Telephone,
			&user.Address,
			&user.City,
			&user.PostalCard,
		)
		if err != nil {
			log.Fatal("Cannot scan the event - ", err)
		}
		users = append(users, user)
	}
	return users
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Fatal("Failed hashing the password - ", err)
	}
	return string(bytes)
}

// Return true if the role corresponding to the session_id given is "superior" to the role provided
func (s *Store) Authorize(session_id string, roleMin types.Role) bool {
	var rolestr string
	query := `
        SELECT role FROM users WHERE id=
        (SELECT userid FROM sessions WHERE id=?);
    `
	err := s.DB.QueryRow(query, session_id).Scan(&rolestr)
	// TODO: do better error handling on that, I do not want to stop the server if that request fails.
	if err != nil {
		log.Fatalf("Failed to find the role of the user referred to the session id %q - %s)", session_id, err)
	}
	userRole := types.ConvertToRole(rolestr)

	return userRole.IsSuperior(roleMin)
}
