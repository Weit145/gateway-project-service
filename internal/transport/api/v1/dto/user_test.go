package dto

import (
	"encoding/json"
	"testing"

	"github.com/Weit145/gateway-project-service/internal/domain/auth"
	"github.com/google/uuid"
)

func TestUserJSONWithoutName(t *testing.T) {
	id := uuid.New()
	user, err := auth.NewUser(id, "owner@example.com", "user")
	if err != nil {
		t.Fatal(err)
	}
	for _, dto := range []User{DomainToUserDTO(user), NewUserDTO(id, user.Email(), user.Role())} {
		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]string
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 3 || fields["id"] != id.String() || fields["email"] != user.Email() || fields["role"] != user.Role() {
			t.Fatalf("unexpected user JSON: %s", data)
		}
	}
}
