package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	config "github.com/rokmonster/ocr/internal/pkg/config/scannerconfig"
	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
	"github.com/rokmonster/ocr/internal/pkg/rokocr"
	log "github.com/sirupsen/logrus"
)

var flags = config.Parse()

func printResultsTable(data []schema.OCRResponse, template schema.RokOCRTemplate) {
	headers := []string{"Filename"}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	for _, row := range data {
		rowData := []string{row.Filename}

		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row.Data[x.Field]))
		}
		table.Append(rowData)
	}

	table.Render()
}

func writeCSV(data []schema.OCRResponse, template schema.RokOCRTemplate) {

	fd, err := os.Create(fmt.Sprintf("%s/%v.csv", flags.OutputDirectory, time.Now().Unix()))
	if err != nil {
		log.Fatalf("Failed to write csv: %v", err)
		return
	}
	defer fd.Close()

	rokocr.WriteCSV(data, template, fd)
}

func main() {
	rokocr.Prepare(flags.CommonConfiguration)

	force := false
	var template schema.RokOCRTemplate

	if len(strings.TrimSpace(flags.ForceTemplate)) > 0 {
		force = true
		template, _ = schema.LoadTemplate(flags.ForceTemplate)
		log.Infof("Running scanner in force mode with template: %v (%vx%v)", template.Title, template.Width, template.Height)
	} else {
		templates := rokocr.LoadTemplates(flags.TemplatesDirectory)
		if len(templates) == 0 {
			log.Fatalf("No templates found in: %v", flags.TemplatesDirectory)
		}
		log.Debugf("Loaded %v templates", len(templates))
		template = rokocr.FindTemplate(flags.MediaDirectory, templates)
		log.Infof("I think this template is best match: %v (%vx%v)", template.Title, template.Width, template.Height)
	}

	data := rokocr.RunRecognition(flags.MediaDirectory, flags.TessdataDirectory, template, force)

	printResultsTable(data, template)
	writeCSV(data, template)
}
