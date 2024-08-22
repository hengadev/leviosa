package user

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
		name    string
	}{
		{email: "", wantErr: true, name: "Empty email"},
		{email: "awefawe awefawe", wantErr: true, name: "Contain white space"},
		{email: `aw"fwe'awef`, wantErr: true, name: "Contain quote"},
		{
			email:   "awfawefawefawfawefawfwafawfghawjfkwupwfwr9i24r23rfdfwhfwefwaefefweawefwaefwaefawfawefwfawrfwefewafgwefwafawefwafawefawrfawfawefww",
			wantErr: true,
			name:    fmt.Sprintf("Cannot be other %d in lenght", emailMaxLength),
		},
		{email: "misstheatcharacter", wantErr: true, name: "Contain white space"},
		{email: "@misstheatcharacter", wantErr: true, name: "Miss content before the @"},
		{email: "misstheatcharacter@", wantErr: true, name: "Miss content after the @"},
		{email: "John DOE <john.doe@gmail.com>", wantErr: true, name: "Cannnot include a name"},
		{email: "john.doe@gmail", wantErr: true, name: "Missing top level domain"},
		{email: "john.doe@gmail.com", wantErr: false, name: "Good email"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ValidateEmail(tt.email)
			assert.Equal(t, got != nil, tt.wantErr)
		})
	}
}
