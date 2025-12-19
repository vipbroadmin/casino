package player

import "fmt"

// We store enums as ints (SMALLINT) in DB instead of database ENUM.

type Status int16

const (
	StatusUnknown Status = 0
	StatusActive  Status = 1
	StatusBlocked Status = 2
	StatusFrozen  Status = 3
	StatusClosed  Status = 4
)

func (s Status) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusBlocked:
		return "blocked"
	case StatusFrozen:
		return "frozen"
	case StatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func ParseStatus(v string) (Status, error) {
	switch v {
	case "active":
		return StatusActive, nil
	case "blocked":
		return StatusBlocked, nil
	case "frozen":
		return StatusFrozen, nil
	case "closed":
		return StatusClosed, nil
	default:
		return StatusUnknown, fmt.Errorf("%w: %s", ErrInvalidStatus, v)
	}
}

func StatusList() []string {
	return []string{"active", "blocked", "frozen", "closed"}
}

type Gender int16

const (
	GenderNone   Gender = 0
	GenderMale   Gender = 1
	GenderFemale Gender = 2
	GenderOther  Gender = 3
)

func (g Gender) String() string {
	switch g {
	case GenderNone:
		return "none"
	case GenderMale:
		return "male"
	case GenderFemale:
		return "female"
	case GenderOther:
		return "other"
	default:
		return "none"
	}
}

func ParseGender(v string) (Gender, error) {
	switch v {
	case "none":
		return GenderNone, nil
	case "male":
		return GenderMale, nil
	case "female":
		return GenderFemale, nil
	case "other":
		return GenderOther, nil
	default:
		return GenderNone, fmt.Errorf("%w: %s", ErrInvalidGender, v)
	}
}

func GenderList() []string {
	return []string{"none", "male", "female", "other"}
}

type ActorType int16

const (
	ActorPlayer ActorType = 1
	ActorAdmin  ActorType = 2
	ActorSystem ActorType = 3
)

func (a ActorType) String() string {
	switch a {
	case ActorPlayer:
		return "player"
	case ActorAdmin:
		return "administrator"
	case ActorSystem:
		return "system"
	default:
		return "system"
	}
}
