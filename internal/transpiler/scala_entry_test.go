package transpiler

import "testing"

func TestApplicationEntry(t *testing.T) {
	for _, tc := range []struct{ name, code, want string }{
		{"custom name", "object MyApp extends ZIOAppDefault { def run = effect }", "MyApp"},
		{"qualified", "object Demo extends zio.ZIOAppDefault:\n  def run = effect", "Demo"},
		{"comments", "/* object Wrong extends ZIOAppDefault /* nested */ */\nobject MyApp /* note */ extends ZIOAppDefault {}", "MyApp"},
		{"strings", "val example = \"\"\"object Wrong extends ZIOAppDefault {}\"\"\"\nobject MyApp extends ZIOAppDefault {}", "MyApp"},
		{"escaped string", "val example = \"\\\" object Wrong extends ZIOAppDefault {}\"\nobject MyApp extends ZIOAppDefault {}", "MyApp"},
		{"backticks", "object `My App` extends ZIOAppDefault {}", "`My App`"},
		{"char literals", "val quote = '\"'; val escaped = '\\''; object MyApp extends ZIOAppDefault {}", "MyApp"},
		{"plain script", "println(42)", ""},
		{"multiple apps", "object A extends ZIOAppDefault {}\nobject B extends ZIOAppDefault {}", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := applicationEntry(tc.code)
			if tc.want == "" {
				if err == nil {
					t.Fatal("expected invalid application error")
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("entry=%q error=%v, want %q", got, err, tc.want)
			}
		})
	}
}
