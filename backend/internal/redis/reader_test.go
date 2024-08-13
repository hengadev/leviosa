package redis_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/GaryHY/event-reservation-app/internal/domain/session"
	"github.com/GaryHY/event-reservation-app/tests"
)

func TestFindSessionByID(t *testing.T) {
	// TODO: test cases
	// - id not in database
	// - id in database valid classic
	tests := []struct {
		sess        *session.Session
		wantErr     bool
		wantSession bool
		inits       []Init
		name        string
	}{
		{sess: &sessionTest, wantErr: true, wantSession: false, inits: nil, name: "ID not in database"},
		{sess: &sessionTest, wantErr: false, wantSession: true, inits: []Init{{key: sessionTest.ID, value: sessionTest.Values()}}, name: "ID in database"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			repo, err := setupSessionRepo(ctx, t, tt.inits...)
			_, err = initRepo(ctx, repo, tt.inits...)
			if err != nil {
				t.Errorf("init the repo")
			}
			if err != nil {
				t.Errorf("setup database: %s", err)
			}
			sessionRes, err := repo.FindSessionByID(ctx, tt.sess.ID)
			test.Assert(t, err != nil, tt.wantErr)

			if tt.wantSession {
				types := reflect.TypeOf(*sessionRes)
				fmt.Println("the types of sessionRes: ", types)
				gotV := reflect.ValueOf(*tt.sess)
				expectV := reflect.ValueOf(*sessionRes)
				fields := reflect.VisibleFields(types)
				for _, field := range fields {
					if field.Type == reflect.TypeOf(time.Time{}) {
						got := gotV.FieldByName(field.Name)
						expect := expectV.FieldByName(field.Name)
						test.Assert(t, got != expect, tt.wantSession)
					}
				}
			}
		})
	}
}
