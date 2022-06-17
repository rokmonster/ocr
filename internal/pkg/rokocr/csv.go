package rokocr

import (
	"encoding/csv"
	"fmt"
	"io"

	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
)

func WriteCSV(data []schema.OCRResponse, template schema.RokOCRTemplate, w io.Writer) {
	headers := []string{"Filename"}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	table := csv.NewWriter(w)
	table.Write(headers)
	for _, row := range data {
		rowData := []string{row.Filename}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row.Data[x.Field]))
		}
		table.Write(rowData)
	}
	table.Flush()

}
