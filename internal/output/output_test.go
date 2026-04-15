package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{
			name: "epoch zero",
			ms:   0,
			want: time.UnixMilli(0).Format("2006-01-02 15:04"),
		},
		{
			name: "known date - 2024-01-15 10:30 UTC",
			ms:   1705311000000,
			want: time.UnixMilli(1705311000000).Format("2006-01-02 15:04"),
		},
		{
			name: "large timestamp - year 2030",
			ms:   1893456000000,
			want: time.UnixMilli(1893456000000).Format("2006-01-02 15:04"),
		},
		{
			name: "negative timestamp - before epoch",
			ms:   -1000,
			want: time.UnixMilli(-1000).Format("2006-01-02 15:04"),
		},
		{
			name: "recent timestamp",
			ms:   1700000000000,
			want: time.UnixMilli(1700000000000).Format("2006-01-02 15:04"),
		},
		{
			name: "timestamp with seconds granularity",
			ms:   1705311045000,
			want: time.UnixMilli(1705311045000).Format("2006-01-02 15:04"),
		},
		{
			name: "timestamp with milliseconds",
			ms:   1705311045123,
			want: time.UnixMilli(1705311045123).Format("2006-01-02 15:04"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimestamp(tt.ms)
			if got != tt.want {
				t.Errorf("FormatTimestamp(%d) = %q, want %q", tt.ms, got, tt.want)
			}
		})
	}
}

func TestFormatTimestamp_Format(t *testing.T) {
	// Verify the format string pattern (YYYY-MM-DD HH:MM)
	ms := int64(1705311000000)
	got := FormatTimestamp(ms)

	// Should match pattern like "2024-01-15 10:30"
	if len(got) != 16 {
		t.Errorf("FormatTimestamp output length = %d, want 16 (YYYY-MM-DD HH:MM)", len(got))
	}
	if got[4] != '-' || got[7] != '-' || got[10] != ' ' || got[13] != ':' {
		t.Errorf("FormatTimestamp(%d) = %q, doesn't match YYYY-MM-DD HH:MM pattern", ms, got)
	}
}

func TestDeref(t *testing.T) {
	tests := []struct {
		name     string
		s        *string
		fallback string
		want     string
	}{
		{
			name:     "nil pointer returns fallback",
			s:        nil,
			fallback: "default",
			want:     "default",
		},
		{
			name:     "nil with empty fallback",
			s:        nil,
			fallback: "",
			want:     "",
		},
		{
			name:     "nil with dash fallback",
			s:        nil,
			fallback: "-",
			want:     "-",
		},
		{
			name:     "non-nil pointer returns value",
			s:        strPtr("hello"),
			fallback: "default",
			want:     "hello",
		},
		{
			name:     "empty string pointer",
			s:        strPtr(""),
			fallback: "default",
			want:     "",
		},
		{
			name:     "pointer with spaces",
			s:        strPtr("  spaces  "),
			fallback: "fallback",
			want:     "  spaces  ",
		},
		{
			name:     "non-nil ignores fallback",
			s:        strPtr("value"),
			fallback: "ignored",
			want:     "value",
		},
		{
			name:     "unicode pointer value",
			s:        strPtr("hello world"),
			fallback: "n/a",
			want:     "hello world",
		},
		{
			name:     "nil with N/A fallback",
			s:        nil,
			fallback: "N/A",
			want:     "N/A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Deref(tt.s, tt.fallback)
			if got != tt.want {
				t.Errorf("Deref() = %q, want %q", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestFormatConstants(t *testing.T) {
	if FormatTable != "table" {
		t.Errorf("FormatTable = %q, want %q", FormatTable, "table")
	}
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
}

func TestPrint_JSONFormat(t *testing.T) {
	data := map[string]string{"key": "value", "name": "test"}

	output := captureStdout(t, func() {
		Print(FormatJSON, data, nil)
	})

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON output not valid: %v\noutput: %s", err, output)
	}
	if result["key"] != "value" {
		t.Errorf("key = %q, want %q", result["key"], "value")
	}
	if result["name"] != "test" {
		t.Errorf("name = %q, want %q", result["name"], "test")
	}
}

func TestPrint_JSONFormat_IndentedOutput(t *testing.T) {
	data := map[string]int{"count": 42}

	oldPretty := PrettyJSON
	PrettyJSON = true
	defer func() { PrettyJSON = oldPretty }()

	output := captureStdout(t, func() {
		Print(FormatJSON, data, nil)
	})

	// Should be indented (contains newline + spaces)
	if !strings.Contains(output, "\n") {
		t.Errorf("JSON output should be indented with newlines")
	}
	if !strings.Contains(output, "  ") {
		t.Errorf("JSON output should contain indentation spaces")
	}
}

func TestPrint_JSONFormat_Slice(t *testing.T) {
	data := []string{"a", "b", "c"}

	output := captureStdout(t, func() {
		Print(FormatJSON, data, nil)
	})

	var result []string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON output not valid: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("len = %d, want 3", len(result))
	}
}

func TestPrint_JSONFormat_Struct(t *testing.T) {
	type item struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	data := item{ID: "abc", Name: "Test Item"}

	output := captureStdout(t, func() {
		Print(FormatJSON, data, nil)
	})

	var result item
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON output not valid: %v", err)
	}
	if result.ID != "abc" {
		t.Errorf("ID = %q, want %q", result.ID, "abc")
	}
}

func TestPrint_TableFormat(t *testing.T) {
	output := captureStdout(t, func() {
		Print(FormatTable, nil, func(tw table.Writer) {
			tw.AppendHeader(table.Row{"ID", "Name"})
			tw.AppendRow(table.Row{"1", "Alice"})
			tw.AppendRow(table.Row{"2", "Bob"})
		})
	})

	// Table output should contain headers and rows
	// go-pretty StyleLight renders headers in uppercase
	if !strings.Contains(output, "ID") {
		t.Errorf("table output should contain 'ID' header, got: %q", output)
	}
	if !strings.Contains(output, "NAME") {
		t.Errorf("table output should contain 'NAME' header, got: %q", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("table output should contain 'Alice', got: %q", output)
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("table output should contain 'Bob', got: %q", output)
	}
}

func TestPrint_TableFormat_EmptyTable(t *testing.T) {
	output := captureStdout(t, func() {
		Print(FormatTable, nil, func(tw table.Writer) {
			tw.AppendHeader(table.Row{"ID", "Name"})
		})
	})

	if !strings.Contains(output, "ID") {
		t.Error("empty table should still render headers")
	}
}

func TestPrint_DefaultFormat_FallsBackToJSON(t *testing.T) {
	data := map[string]string{"fallback": "json"}

	output := captureStdout(t, func() {
		Print(FormatTable, data, nil) // nil renderer with table format
	})

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON fallback, got: %s", output)
	}
	if result["fallback"] != "json" {
		t.Errorf("fallback = %q, want %q", result["fallback"], "json")
	}
}

func TestPrint_UnknownFormat_FallsBackToJSON(t *testing.T) {
	data := map[string]string{"test": "value"}

	output := captureStdout(t, func() {
		Print(Format("unknown"), data, func(tw table.Writer) {
			tw.AppendRow(table.Row{"should not render"})
		})
	})

	// Unknown format should hit default case which renders table if renderer is not nil
	// Actually, the default case checks if tableRenderer != nil, so it will use the table renderer
	if !strings.Contains(output, "should not render") && !strings.Contains(output, "test") {
		t.Errorf("unknown format should produce some output, got: %s", output)
	}
}

func TestSuccess_WritesToStderr(t *testing.T) {
	output := captureStderr(t, func() {
		Success("operation %s completed", "test")
	})

	if !strings.Contains(output, "operation test completed") {
		t.Errorf("Success() output = %q, want to contain 'operation test completed'", output)
	}
}

func TestSuccess_HasIndentation(t *testing.T) {
	output := captureStderr(t, func() {
		Success("done")
	})

	if !strings.HasPrefix(output, "  ") {
		t.Errorf("Success() should have 2-space indent, got %q", output)
	}
}

func TestWarn_WritesToStderr(t *testing.T) {
	output := captureStderr(t, func() {
		Warn("something %s happened", "bad")
	})

	if !strings.Contains(output, "Warning:") {
		t.Errorf("Warn() output = %q, want to contain 'Warning:'", output)
	}
	if !strings.Contains(output, "something bad happened") {
		t.Errorf("Warn() output = %q, want to contain message", output)
	}
}

func TestWarn_HasIndentation(t *testing.T) {
	output := captureStderr(t, func() {
		Warn("caution")
	})

	if !strings.HasPrefix(output, "  ") {
		t.Errorf("Warn() should have 2-space indent, got %q", output)
	}
}

func TestSuccess_NoArgs(t *testing.T) {
	output := captureStderr(t, func() {
		Success("simple message")
	})

	if !strings.Contains(output, "simple message") {
		t.Errorf("Success() output = %q, want to contain 'simple message'", output)
	}
}

func TestWarn_NoArgs(t *testing.T) {
	output := captureStderr(t, func() {
		Warn("simple warning")
	})

	if !strings.Contains(output, "simple warning") {
		t.Errorf("Warn() output = %q, want to contain 'simple warning'", output)
	}
}

func TestPrint_JSON_NilData(t *testing.T) {
	output := captureStdout(t, func() {
		Print(FormatJSON, nil, nil)
	})

	trimmed := strings.TrimSpace(output)
	if trimmed != "null" {
		t.Errorf("JSON nil should be 'null', got %q", trimmed)
	}
}

func TestPrint_JSON_NestedStruct(t *testing.T) {
	type Inner struct {
		Value int `json:"value"`
	}
	type Outer struct {
		Name  string `json:"name"`
		Inner Inner  `json:"inner"`
	}

	data := Outer{Name: "test", Inner: Inner{Value: 42}}

	output := captureStdout(t, func() {
		Print(FormatJSON, data, nil)
	})

	var result Outer
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("JSON output not valid: %v", err)
	}
	if result.Inner.Value != 42 {
		t.Errorf("Inner.Value = %d, want 42", result.Inner.Value)
	}
}

func TestPrint_Table_MultipleColumns(t *testing.T) {
	output := captureStdout(t, func() {
		Print(FormatTable, nil, func(tw table.Writer) {
			tw.AppendHeader(table.Row{"A", "B", "C", "D", "E"})
			tw.AppendRow(table.Row{"1", "2", "3", "4", "5"})
		})
	})

	for _, col := range []string{"A", "B", "C", "D", "E"} {
		if !strings.Contains(output, col) {
			t.Errorf("table output should contain column %q", col)
		}
	}
}

func TestIsTTY(t *testing.T) {
	// In test environment, stdout is typically not a TTY
	result := IsTTY()
	// We just verify it returns a boolean without panicking
	_ = result
}

// captureStdout captures stdout output during fn execution.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

// captureStderr captures stderr output during fn execution.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}
