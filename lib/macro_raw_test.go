package lib

import (
	"strings"
	"testing"
)

const infoMacro = `<ac:structured-macro ac:name="info"><ac:rich-text-body><p>Hi</p></ac:rich-text-body></ac:structured-macro>`

func TestRawMacroBlockIsWrittenAsIs(t *testing.T) {
	in := "Text\n\n```CONFLUENCE-MACRO-RAW\n" + infoMacro + "\n```\n"

	out, _, err := renderContent("page.md", in, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, infoMacro) {
		t.Fatalf("expected the raw macro in the output, got:\n%s", out)
	}
	if strings.Contains(out, `ac:name="code"`) || strings.Contains(out, "CDATA") {
		t.Fatalf("the raw macro must not be wrapped in a code macro, got:\n%s", out)
	}
}

func TestRegularCodeBlockIsStillACodeMacro(t *testing.T) {
	in := "```go\nfmt.Println(1)\n```\n"

	out, _, err := renderContent("page.md", in, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `<ac:structured-macro ac:name="code"`) || !strings.Contains(out, "fmt.Println(1)") {
		t.Fatalf("expected a code macro, got:\n%s", out)
	}
}
