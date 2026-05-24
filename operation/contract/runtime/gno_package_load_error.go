package runtime

import (
	"fmt"
	"strings"
	"unicode/utf8"

	gstore "github.com/gnolang/gno/tm2/pkg/store"
)

const maxGnoPackageLoadCauseBytes = 512

func newGnoPackageLoadPanicError(r any) error {
	cause := sanitizeGnoPackageLoadPanic(r)
	if cause == "" {
		return fmt.Errorf("package load failed")
	}

	return fmt.Errorf("%s", cause)
}

func isGnoPackageLoadResourceLimitPanic(r any, gasMeter gstore.GasMeter) bool {
	if gasMeter != nil && gasMeter.IsOutOfGas() {
		return true
	}

	msg := strings.ToLower(fmt.Sprint(r))
	return strings.Contains(msg, "alloc") || strings.Contains(msg, "allocation")
}

// sanitizeGnoPackageLoadPanic keeps the first compact diagnostic emitted while
// loading/typechecking contract code, but omits multi-line VM rendering and
// stack output. Execution-time panics do not use this path.
func sanitizeGnoPackageLoadPanic(r any) string {
	raw := strings.ReplaceAll(fmt.Sprint(r), "\r\n", "\n")
	for _, line := range strings.Split(raw, "\n") {
		line = strings.Join(strings.Fields(line), " ")
		if line == "" || isGnoPackageLoadDumpLine(line) {
			continue
		}

		return truncateGnoPackageLoadCause(line)
	}

	return ""
}

func isGnoPackageLoadDumpLine(line string) bool {
	lower := strings.ToLower(line)

	return strings.HasPrefix(lower, "---") ||
		strings.HasPrefix(lower, "goroutine ") ||
		strings.HasPrefix(lower, "stack trace") ||
		strings.HasPrefix(lower, "github.com/gnolang/") ||
		strings.Contains(lower, "/gnovm/pkg/") ||
		strings.HasPrefix(lower, "runtime.")
}

func truncateGnoPackageLoadCause(cause string) string {
	if len(cause) <= maxGnoPackageLoadCauseBytes {
		return cause
	}

	end := maxGnoPackageLoadCauseBytes - len("...")
	for end > 0 && !utf8.RuneStart(cause[end]) {
		end--
	}

	return cause[:end] + "..."
}
