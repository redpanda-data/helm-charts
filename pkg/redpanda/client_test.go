package redpanda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFirstUser(t *testing.T) {
	cases := []struct {
		In  string
		Out [3]string
	}{
		{
			In:  "hello:world:SCRAM-SHA-256",
			Out: [3]string{"hello", "world", "SCRAM-SHA-256"},
		},
		{
			In:  "name:password\n#Intentionally Blank\n",
			Out: [3]string{"name", "password", "SCRAM-SHA-512"},
		},
		{
			In:  "name:password:SCRAM-MD5-999",
			Out: [3]string{"", "", ""},
		},
	}

	for _, c := range cases {
		user, password, mechanism := firstUser([]byte(c.In))
		assert.Equal(t, [3]string{user, password, mechanism}, c.Out)
	}
}
