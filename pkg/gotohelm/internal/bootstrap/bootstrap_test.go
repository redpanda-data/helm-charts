//go:build !gotohelm

package bootstrap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestTime(t *testing.T) {
	// Any duration with a minimum of second precision can be parsed just like go:
	t.Run("parse duration", rapid.MakeCheck(func(t *rapid.T) {
		// Any duration that is a round number of seconds.
		dur := time.Duration(rapid.Uint64Min(uint64(time.Second)).Filter(func(u uint64) bool {
			return u%uint64(time.Second) == 0
		}).Draw(t, "duration"))

		out := time_ParseDuration(dur.String())

		// time.ParseDuration == time_ParseDuration
		assert.Equal(t, int64(dur), out)
	}))

	// Any duration with a minimum of second precision can be stringified just like go:
	t.Run("string", rapid.MakeCheck(func(t *rapid.T) {
		// Any duration that is a round number of seconds.
		dur := time.Duration(rapid.Uint64Min(uint64(time.Second)).Filter(func(u uint64) bool {
			return u%uint64(time.Second) == 0
		}).Draw(t, "duration"))

		// time.Duration.String == time_Duration_String
		assert.Equal(t, dur.String(), time_Duration_String(int64(dur)))
	}))

	// Durations < 1s and > 0 fail to parse
	t.Run("small durations", rapid.MakeCheck(func(t *rapid.T) {
		dur := time.Duration(rapid.Uint64Range(1, uint64(time.Second)-1).Draw(t, "duration"))

		assert.Panicsf(t, func() {
			time_ParseDuration(dur.String())
		}, "%q should cause a panic", dur.String())
	}))

	// Strings that fail on time.ParseDuration should fail (via panic) on time_ParseDuration as well.
	t.Run("invalid durations", rapid.MakeCheck(func(t *rapid.T) {
		s := rapid.String().Filter(func(s string) bool {
			// Only pull invalid durations.
			_, err := time.ParseDuration(s)
			return err != nil
		}).Draw(t, "in")

		assert.Panics(t, func() {
			_ = time_ParseDuration(s)
		}, "%q failed time.ParseDuration but time_ParseDuration didn't panic", s)
	}))
}
