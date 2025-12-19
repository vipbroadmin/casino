package player

import (
	"fmt"
	"regexp"
	"time"
)

type Address struct {
	CountryCode string
	Locale      string
	TimeZone    string
}

var (
	reCountry = regexp.MustCompile(`^[A-Z]{2}$`)
	reLocale  = regexp.MustCompile(`^[a-z]{2}([_-][A-Z]{2})?$`)
)

func NewAddress(countryCode, locale, timeZone string) (Address, error) {
	a := Address{
		CountryCode: countryCode,
		Locale:      locale,
		TimeZone:    timeZone,
	}
	if err := a.Validate(); err != nil {
		return Address{}, err
	}
	return a, nil
}

func (a Address) Validate() error {
	if a.CountryCode != "" && !reCountry.MatchString(a.CountryCode) {
		return fmt.Errorf("%w: %s", ErrInvalidCountryCode, a.CountryCode)
	}
	if a.Locale != "" && !reLocale.MatchString(a.Locale) {
		return fmt.Errorf("%w: %s", ErrInvalidLocale, a.Locale)
	}
	if a.TimeZone != "" {
		if _, err := time.LoadLocation(a.TimeZone); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidTimeZone, a.TimeZone)
		}
	}
	return nil
}
