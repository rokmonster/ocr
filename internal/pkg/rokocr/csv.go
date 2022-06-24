package rokocr

import (
	"encoding/csv"
	"fmt"
	"io"

	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
)

func WriteCSV(data []schema.OCRResult, template schema.OCRTemplate, w io.Writer) {
	headers := []string{"Filename"}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	table := csv.NewWriter(w)
	_ = table.Write(headers)
	for _, row := range data {
		rowData := []string{row.Filename}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row.Data[x.Field]))
		}
		_ = table.Write(rowData)
	}
	table.Flush()

}
