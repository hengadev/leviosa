package userRepository

import (
	"context"
	"database/sql"

	"github.com/GaryHY/event-reservation-app/internal/domain/user"

	_ "github.com/mattn/go-sqlite3"
)

type repository struct {
	DB *sql.DB
}

func (u *repository) GetDB() *sql.DB {
	return u.DB
}

func New(ctx context.Context, db *sql.DB) *repository {
	return &repository{db}
}

// NOTE: the basic api for the use of the user repo

type Readerrepo struct {
	DB *sql.DB
}

func Newreaderrepo(db *sql.DB) *Readerrepo {
	return &Readerrepo{db}
}

func (r *Readerrepo) FindAccountByID(ctx context.Context, id int) (*user.User, error) {
	return nil, nil
}
func (r *Readerrepo) GetCredentials(ctx context.Context, usr *user.Credentials) (int, string, user.Role, error) {
	return 0, "", user.UNKNOWN, nil
}

// TODO: make a new reader repo for that one
type otherrepo struct {
	*Readerrepo
}

func (o *otherrepo) AddAccount(ctx context.Context, user *user.User) error {
	return nil
}
func (o *otherrepo) ModifyAccount(ctx context.Context, user *user.User, whereMap map[string]any, prohibitedFields ...string) error {
	return nil
}

func Newotherrepo(rp *Readerrepo) *otherrepo {
	return &otherrepo{rp}
}
