package csvio

import (
	"bytes"
	"strings"
	"testing"
)

type testProduct struct {
	ID    string `csv:"id"`
	Name  string `csv:"name"`
	Price int    `csv:"price"`
}

type testMinimal struct {
	Key string `csv:"key"`
}

func TestExportCSV_RoundTrip(t *testing.T) {
	items := []testProduct{
		{ID: "prod_1", Name: "Monthly", Price: 999},
		{ID: "prod_2", Name: "Annual", Price: 4999},
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, items); err != nil {
		t.Fatalf("ExportCSV() error: %v", err)
	}

	got, err := ImportCSV[testProduct](strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ImportCSV() error: %v", err)
	}

	if len(got) != len(items) {
		t.Fatalf("ImportCSV() returned %d items, want %d", len(got), len(items))
	}
	for i, item := range got {
		if item != items[i] {
			t.Errorf("item[%d] = %+v, want %+v", i, item, items[i])
		}
	}
}

func TestExportCSV_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := ExportCSV(&buf, []testProduct{})
	if err != nil {
		t.Fatalf("ExportCSV(empty) error: %v", err)
	}

	// Empty slice should produce no output (no header, no rows)
	if buf.Len() != 0 {
		t.Errorf("ExportCSV(empty) wrote %d bytes, want 0", buf.Len())
	}
}

func TestExportCSV_SingleItem(t *testing.T) {
	items := []testProduct{
		{ID: "prod_1", Name: "Solo", Price: 100},
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, items); err != nil {
		t.Fatalf("ExportCSV() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "id") {
		t.Error("CSV output should contain header 'id'")
	}
	if !strings.Contains(output, "prod_1") {
		t.Error("CSV output should contain value 'prod_1'")
	}
}

func TestExportCSV_MinimalStruct(t *testing.T) {
	items := []testMinimal{
		{Key: "abc"},
		{Key: "def"},
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, items); err != nil {
		t.Fatalf("ExportCSV() error: %v", err)
	}

	got, err := ImportCSV[testMinimal](strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ImportCSV() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("ImportCSV() returned %d items, want 2", len(got))
	}
	if got[0].Key != "abc" {
		t.Errorf("item[0].Key = %q, want %q", got[0].Key, "abc")
	}
	if got[1].Key != "def" {
		t.Errorf("item[1].Key = %q, want %q", got[1].Key, "def")
	}
}

func TestImportCSV_InvalidCSV(t *testing.T) {
	// Completely empty input - no header at all
	_, err := ImportCSV[testProduct](strings.NewReader(""))
	if err == nil {
		t.Error("ImportCSV(empty) should return error for missing header")
	}
}

func TestImportCSV_HeaderOnly(t *testing.T) {
	csv := "id,name,price\n"
	got, err := ImportCSV[testProduct](strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportCSV(header-only) error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("ImportCSV(header-only) returned %d items, want 0", len(got))
	}
}

func TestImportCSV_ExtraColumns(t *testing.T) {
	csv := "id,name,price,extra\nprod_1,Monthly,999,ignored\n"
	got, err := ImportCSV[testProduct](strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ImportCSV(extra columns) error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ImportCSV(extra columns) returned %d items, want 1", len(got))
	}
	if got[0].ID != "prod_1" {
		t.Errorf("item.ID = %q, want %q", got[0].ID, "prod_1")
	}
}

func TestExportCSV_SpecialCharacters(t *testing.T) {
	items := []testProduct{
		{ID: "prod_1", Name: "Monthly, Premium", Price: 999},
		{ID: "prod_2", Name: `Annual "Pro"`, Price: 4999},
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, items); err != nil {
		t.Fatalf("ExportCSV() error: %v", err)
	}

	got, err := ImportCSV[testProduct](strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ImportCSV() error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("ImportCSV() returned %d items, want 2", len(got))
	}
	if got[0].Name != "Monthly, Premium" {
		t.Errorf("item[0].Name = %q, want %q", got[0].Name, "Monthly, Premium")
	}
	if got[1].Name != `Annual "Pro"` {
		t.Errorf("item[1].Name = %q, want %q", got[1].Name, `Annual "Pro"`)
	}
}

func TestExportCSV_ManyItems(t *testing.T) {
	items := make([]testProduct, 100)
	for i := range items {
		items[i] = testProduct{
			ID:    "prod_" + strings.Repeat("x", i%10),
			Name:  "Product",
			Price: i,
		}
	}

	var buf bytes.Buffer
	if err := ExportCSV(&buf, items); err != nil {
		t.Fatalf("ExportCSV(100 items) error: %v", err)
	}

	got, err := ImportCSV[testProduct](strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("ImportCSV(100 items) error: %v", err)
	}
	if len(got) != 100 {
		t.Errorf("ImportCSV() returned %d items, want 100", len(got))
	}
}
