package source

import (
	"errors"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidID  = errors.New("invalid source id")
	validIDRegExp = regexp.MustCompile("[a-z0-9-]+")
)

type Source struct {
	ID              string `yaml:"id"`
	Bucket          string `yaml:"bucket"`
	Prefix          string `yaml:"prefix"`
	LookbackEntries int    `yaml:"lookback-entries"`
	AliasID         bool   `yaml:"alias-id"`
}

func validateID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: id must be set", ErrInvalidID)
	}
	if testID := validIDRegExp.FindString(id); id != testID {
		return fmt.Errorf("%w: id must match the regex /%s/", ErrInvalidID, validIDRegExp.String())
	}
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *Source) UnmarshalYAML(value *yaml.Node) error {
	type RawSource Source
	raw := &RawSource{}
	if err := value.Decode(raw); err != nil {
		return err
	}
	if err := validateID(raw.ID); err != nil {
		return err
	}
	*s = Source(*raw)
	return nil
}