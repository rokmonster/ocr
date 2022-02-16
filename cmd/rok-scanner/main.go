package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	config "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config/scannerconfig"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	rokocr "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
)

var flags = config.Parse()

func printResultsTable(data []schema.OCRResponse, template *schema.RokOCRTemplate) {
	headers := []string{}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	for _, row := range data {
		rowData := []string{}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row[x.Field]))
		}
		table.Append(rowData)
	}

	table.Render()
}

func writeCSV(data []schema.OCRResponse, template *schema.RokOCRTemplate) {
	headers := []string{}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	fd, err := os.Create(fmt.Sprintf("%s/%v.csv", flags.OutputDirectory, time.Now().Unix()))
	if err != nil {
		log.Fatalf("Failed to write csv: %v", err)
		return
	}
	defer fd.Close()

	table := csv.NewWriter(fd)
	table.Write(headers)
	for _, row := range data {
		rowData := []string{}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row[x.Field]))
		}
		table.Write(rowData)
	}
	table.Flush()

}

func main() {
	templates := rokocr.LoadTemplates(flags.TemplatesDirectory)
	if len(templates) == 0 {
		log.Fatalf("No templates found in: %v", flags.TemplatesDirectory)
	}

	log.Debugf("Loaded %v templates", len(templates))

	template := rokocr.FindTemplate(flags.MediaDirectory, templates)
	data := rokocr.RunRecognition(flags.MediaDirectory, flags.TessdataDirectory, template)

	printResultsTable(data, template)
	writeCSV(data, template)
}
