package runtime

import "testing"

func TestAnalyzeContractSchemaAllowedImports(t *testing.T) {
	source := `package contract
import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"mitum/chain"
	"strconv"
	"strings"
	"unicode/utf8"
)

var text string

func Initialize(ctx chain.WriteContext) error {
	buf := bytes.NewBufferString(ctx.GetSender())
	text = strings.ToUpper(buf.String())
	text = text + ":" + strconv.FormatInt(ctx.GetHeight(), 10)
	text = text + ":" + hex.EncodeToString([]byte{1, 2})
	text = text + ":" + base64.StdEncoding.EncodeToString([]byte{3, 4})
	if !utf8.ValidString(text) {
		return errors.New("invalid utf8")
	}
	return nil
}
`

	if _, err := AnalyzeContractSchema(source); err != nil {
		t.Fatalf("AnalyzeContractSchema returned error: %v", err)
	}
}

func TestAnalyzeContractSchemaRejectsMathRandImport(t *testing.T) {
	source := `package contract
import (
	"math/rand"
	"mitum/chain"
)

func Initialize(ctx chain.WriteContext) error {
	_ = rand.Int()
	return nil
}
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected math/rand import rejection")
	}
	if got := err.Error(); got == "" || !containsAll(got, `import "math/rand" is not allowed`, "allowed imports are") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRejectsFmtImport(t *testing.T) {
	source := `package contract
import (
	"fmt"
	"mitum/chain"
)

func Initialize(ctx chain.WriteContext) error {
	_ = fmt.Sprintf("%s", ctx.GetSender())
	return nil
}
`

	_, err := AnalyzeContractSchema(source)
	if err == nil {
		t.Fatalf("expected fmt import rejection")
	}
	if got := err.Error(); got == "" || !containsAll(got, `import "fmt" is not allowed`, "allowed imports are") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAnalyzeContractSchemaRejectsOtherDisallowedImports(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "time", path: "time"},
		{name: "regexp", path: "regexp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := `package contract
import (
	"` + tt.path + `"
	"mitum/chain"
)

func Initialize(ctx chain.WriteContext) error {
	return nil
}
`

			_, err := AnalyzeContractSchema(source)
			if err == nil {
				t.Fatalf("expected %s import rejection", tt.path)
			}
			if got := err.Error(); got == "" || !containsAll(got, `import "`+tt.path+`" is not allowed`, "allowed imports are") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
