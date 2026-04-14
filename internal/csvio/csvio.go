package csvio

import (
	"encoding/csv"
	"io"

	"github.com/jszwec/csvutil"
)

// ExportCSV writes items as CSV to w.
func ExportCSV[T any](w io.Writer, items []T) error {
	csvW := csv.NewWriter(w)
	enc := csvutil.NewEncoder(csvW)
	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return err
		}
	}
	csvW.Flush()
	return csvW.Error()
}

// ImportCSV reads CSV from r and returns a slice of items.
func ImportCSV[T any](r io.Reader) ([]T, error) {
	csvR := csv.NewReader(r)
	dec, err := csvutil.NewDecoder(csvR)
	if err != nil {
		return nil, err
	}
	var items []T
	for {
		var item T
		if err := dec.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
