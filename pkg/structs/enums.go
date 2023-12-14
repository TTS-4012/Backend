package structs

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
	return "??"
}
