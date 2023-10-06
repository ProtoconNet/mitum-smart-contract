package digest

import (
	"fmt"
	"strings"

	"github.com/ProtoconNet/mitum2/util/hint"
	"github.com/pkg/errors"
)

const (
	ProblemMimetype    = "application/problem+json; charset=utf-8"
	ProblemNamespace   = "https://github.com/ProtoconNet/mitum-currency/v3/problems"
	DefaultProblemType = "others"
)

var (
	ProblemHint = hint.MustNewHint("mitum-currency-problem-v0.0.1")
)

// Problem implements "Problem Details for HTTP
// APIs"<https://tools.ietf.org/html/rfc7807>.
type Problem struct {
	t      string // NOTE http problem type
	title  string
	detail string
	extra  map[string]interface{}
}

func NewProblem(t, title string) Problem {
	return Problem{t: t, title: title}
}

func NewProblemFromError(err error) Problem {
	title, detail := makeSplitedError(err)
	return Problem{
		t:      DefaultProblemType,
		title:  fmt.Sprintf("%s", title),
		detail: fmt.Sprintf("%+v", detail),
	}
}

func (Problem) Hint() hint.Hint {
	return ProblemHint
}

func (pr Problem) Error() string {
	return pr.title
}

func (pr Problem) SetTitle(title string) Problem {
	pr.title = title

	return pr
}

func (pr Problem) SetDetail(detail string) Problem {
	pr.detail = detail

	return pr
}

func (pr Problem) SetExtra(key string, value interface{}) Problem {
	if pr.extra == nil {
		pr.extra = map[string]interface{}{}
	}

	pr.extra[key] = value

	return pr
}

func makeProblemNamespace(t string) string {
	return fmt.Sprintf("%s/%s", ProblemNamespace, t)
}

func parseProblemNamespace(s string) (string, error) {
	if !strings.HasPrefix(s, ProblemNamespace) {
		return "", errors.Errorf("invalid problem namespace: %v", s)
	}
	return s[len(ProblemNamespace)+1:], nil
}

func makeSplitedError(err error) (title, detail string) {
	if len(err.Error()) < 1 {
		return "", ""
	}
	errorSlice := strings.Split(err.Error(), ";")
	switch {
	case len(errorSlice) > 2:
		return strings.Join(errorSlice[:len(errorSlice)], ""), errorSlice[len(errorSlice)-1]
	case len(errorSlice) < 2:
		return errorSlice[0], errorSlice[0]
	default:
		return strings.TrimSpace(err.Error()), strings.TrimSpace(errorSlice[1])
	}
}
