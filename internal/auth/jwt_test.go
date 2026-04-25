package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
)

func TestCreateJWT(t *testing.T) {
	userID, err := uuid.Parse("0279e867-922a-4beb-a813-c24a2e4df890")

	if err != nil {
		t.Errorf("Error creating uuid, lol. %v", err)
	}

	cases := []struct {
		input		string
		expected	error
	} {
		{
			input:		"1f09d01f250ea35fb0da443f986401697884398cc7ff2046e4c90527e2b2ae33",
			expected:	nil,
		},
	}

	for _, c := range cases {
		_, actual := MakeJWT(userID, c.input, time.Duration(10 * time.Second))
		if c.expected != actual {
			t.Errorf("Token was not successfully generated: %v", actual)
		}
	}
}
