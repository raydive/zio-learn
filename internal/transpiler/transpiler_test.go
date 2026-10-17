package transpiler

import (
	"strings"
	"testing"
)

func TestTranspile(t *testing.T) {
	compiler := New("3.3.7", []string{"3.3.7", "2.13.18"}, []string{"dev.zio::zio:2.1.27", "dev.zio::zio-streams:2.1.27", "dev.zio::zio-json:1.1.0"})
	code := "object MyApp extends zio.ZIOAppDefault { def run = zio.Console.printLine(42) }"
	for _, version := range []string{"", "3.3.7", "2.13.18"} {
		input, err := compiler.Validate(Input{Code: code, ScalaVersion: version})
		if err != nil {
			t.Fatal(err)
		}
		script := compiler.Transpile(input)
		if version == "" {
			version = "3.3.7"
		}
		for _, want := range []string{"//> using scala " + version, "dev.zio::zio:2.1.27", "dev.zio::zio-streams:2.1.27", "dev.zio::zio-json:1.1.0", "\n" + code + "\n"} {
			if !strings.Contains(script.Source, want) {
				t.Errorf("missing %q in %s", want, script.Source)
			}
		}
		if script.Filename != "playground.sc" || !strings.HasSuffix(script.Source, "\n\nMyApp.main(Array.empty[String])\n") {
			t.Fatalf("unexpected script: %+v", script)
		}
	}
	for _, input := range []Input{{Code: " "}, {Code: "println(42)"}, {Code: code, ScalaVersion: "9.9.9"}} {
		if _, err := compiler.Validate(input); err == nil {
			t.Errorf("accepted invalid input: %+v", input)
		}
	}
}
