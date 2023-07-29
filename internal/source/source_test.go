package source_test

import (
	"bytes"
	"errors"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/khulnasoft-labs/infected-packages/internal/source"
)

func TestParseSources_InvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "empty",
			id:   "",
		},
		{
			name: "invalid underscore",
			id:   "not_allowed",
		},
		{
			name: "invalid space",
			id:   "not allowed",
		},
		{
			name: "invalid many",
			id:   "~!@#$%^&*()_+",
		},
		{
			name: "upper case",
			id:   "This-is-invalid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dec := yaml.NewDecoder(bytes.NewBuffer([]byte("id: " + test.id + "\n")))
			var s *source.Source
			err := dec.Decode(&s)
			if err == nil || !errors.Is(err, source.ErrInvalidID) {
				t.Errorf("Decode() = %v; want %v", err, source.ErrInvalidID)
			}
		})
	}
}

func TestParseSources_ValidIDs(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "single character",
			id:   "a",
		},
		{
			name: "characters only",
			id:   "thisisavalidid",
		},
		{
			name: "dashes",
			id:   "this-is-a-valid-id",
		},
		{
			name: "digits only",
			id:   "1337",
		},
		{
			name: "beings with dash",
			id:   "-valid",
		},
		{
			name: "dashes only",
			id:   "-----",
		},
		{
			name: "all characters",
			id:   "this-is-4-v4l1d-id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dec := yaml.NewDecoder(bytes.NewBuffer([]byte("id: " + test.id + "\n")))
			var s *source.Source
			err := dec.Decode(&s)
			if err != nil {
				t.Errorf("Decode() = %v; want no error", err)
			}
		})
	}
}

func TestInvalidSource(t *testing.T) {
	dec := yaml.NewDecoder(bytes.NewBuffer([]byte("id: a\nlookback-entries: six")))
	var s *source.Source
	err := dec.Decode(&s)
	if err == nil || errors.Is(err, source.ErrInvalidID) {
		t.Errorf("Decode() = %v; want an error", err)
	}
}