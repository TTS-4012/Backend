package structs

import (
	"github.com/ocontest/backend/pkg"
)

// Verdict is the result of a testcase
type Verdict int

const (
	VerdictOK Verdict = 1 << iota
	VerdictWrong
	VerdictTimeLimit
	VerdictMemoryLimit
	VerdictRuntimeError
	VerdictUnknown
	VerdictCompileError
)

func (v Verdict) String() string {
	switch v {
	case VerdictOK:
		return "OK"
	case VerdictWrong:
		return "WR"
	case VerdictTimeLimit:
		return "TL"
	case VerdictMemoryLimit:
		return "ML"
	case VerdictRuntimeError:
		return "RE"
	case VerdictUnknown:
		return "XX"
	case VerdictCompileError:
		return "CE"
	}
	return "XX"
}

func VerdictFromString(s string) Verdict {
	switch s {
	case "OK":
		return VerdictOK
	case "WR":
		return VerdictWrong
	case "TL":
		return VerdictTimeLimit
	case "ML":
		return VerdictMemoryLimit
	case "RE":
		return VerdictRuntimeError
	case "XX":
		return VerdictUnknown
	case "CE":
		return VerdictCompileError
	}
	// TODO: safe error handling
	pkg.Log.Error("unknown verdict", s)
	return -1
}

type RegistrationStatus int

const (
	Owner RegistrationStatus = 1 + iota
	Registered
	NonRegistered
)
