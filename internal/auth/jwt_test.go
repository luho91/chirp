package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	header := http.Header {}

	header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5")

	cases := []struct {
		input		string
		expected	string
	} {
		{
			input:		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5",
			expected:	"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5",	
		},
	}

	for _, c := range cases {
		actual, err := GetBearerToken(header)
		if err != nil {
			t.Errorf("An error happened: %v", err)
		}

		if c.expected != actual {
			t.Errorf("Tokens do not match:\nexpected: %v\n got: %v", c.expected, actual)
		}
	}
}
