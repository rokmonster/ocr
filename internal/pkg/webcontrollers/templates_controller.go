package webcontrollers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/stringutils"
)

type TemplateMakerSession struct {
	imagePath   string
	schema      map[string]schema.ROKOCRSchema
	checkpoints []schema.OCRCheckpoint
}

type TemplatesController struct {
	Router       fiber.Router
	sessions     map[string]TemplateMakerSession
	templatesDir string
	tessdataDir  string
}

func NewTemplatesController(router fiber.Router, templateDir, tessdataDir string) *TemplatesController {
	return &TemplatesController{
		Router:       router,
		sessions:     make(map[string]TemplateMakerSession),
		templatesDir: templateDir,
		tessdataDir:  tessdataDir,
	}
}

// Binding from JSON
type rokTemplateArea struct {
	rokCropCoordinates
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required"`
}

type rokCropCoordinates struct {
	X int `form:"x" json:"x,string" xml:"x"  binding:"required"`
	Y int `form:"y" json:"y,string" xml:"y" binding:"required"`
	W int `form:"w" json:"w,string" xml:"w" binding:"required"`
	H int `form:"h" json:"h,string" xml:"h" binding:"required"`
}

func (controller *TemplatesController) makeTable(s map[string]schema.ROKOCRSchema) []schema.ROKTableField {
	result := []schema.ROKTableField{}

	for k := range s {
		result = append(result, schema.ROKTableField{
			Title: k,
			Field: k,
			Bold:  false,
			Color: "",
		})
	}

	return result
}

func (controller *TemplatesController) buildTemplate(id string, s TemplateMakerSession) *schema.RokOCRTemplate {
	img, _ := imgutils.ReadImage(s.imagePath)
	hash, _ := goimagehash.DifferenceHash(img)

	return &schema.RokOCRTemplate{
		Title:       fmt.Sprintf("ROK OCR Monster Template [%s]", id),
		Version:     "1",
		Fingerprint: fmt.Sprintf("%x", hash.GetHash()),
		Width:       img.Bounds().Dx(),
		Height:      img.Bounds().Dy(),
		Author:      "ROK OCR Template Maker",
		Threshold:   1,
		OCRSchema:   s.schema,
		Table:       controller.makeTable(s.schema),
		Checkpoints: s.checkpoints,
	}
}

func (controller *TemplatesController) Setup() {

	controller.Router.Get("", func(c *fiber.Ctx) error {
		return c.Render("templatemaker_upload", fiber.Map{})
	})

	// create new session, and redirect
	controller.Router.Post("", func(c *fiber.Ctx) error {
		// create session id
		sessionId := stringutils.Random(12)
		// handle file upload...
		file, _ := c.FormFile("image")
		dst := os.TempDir() + "/" + sessionId + filepath.Ext(file.Filename)
		c.SaveFile(file, dst)

		logrus.Debugf("Uploaded file: %s", dst)

		controller.sessions[sessionId] = TemplateMakerSession{
			imagePath:   dst,
			checkpoints: []schema.OCRCheckpoint{},
			schema:      make(map[string]schema.ROKOCRSchema),
		}

		return c.Redirect("/templates/"+sessionId, http.StatusFound)
	})

	controller.Router.Get("/:session", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))
		if _, ok := controller.sessions[sessionId]; ok {
			// check if session exists;
			return c.Render("templatemaker", fiber.Map{
				"sessionId": sessionId,
			})
		}
		return c.Redirect("/templates", http.StatusFound)
	})

	controller.Router.Get("/:session/image", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))
		imagePath := controller.sessions[sessionId].imagePath
		return c.SendFile(imagePath)
	})

	controller.Router.Post("/:session/scan", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))

		if s, ok := controller.sessions[sessionId]; ok {
			img, _ := imgutils.ReadImage(s.imagePath)
			template := controller.buildTemplate(sessionId, s)

			return c.JSON(fiber.Map{
				"fingerprint": fmt.Sprintf("%x", template.Hash().GetHash()),
				"results":     rokocr.ParseImage("test", img, template, os.TempDir(), "./tessdata"),
			})
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{})
	})

	controller.Router.Get("/:session/export", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))

		if s, ok := controller.sessions[sessionId]; ok {
			template := controller.buildTemplate(sessionId, s)
			bytes, _ := json.MarshalIndent(template, "", "    ")
			os.WriteFile(fmt.Sprintf("%s/builder_%s.json", controller.templatesDir, sessionId), bytes, 0644)
			return c.Redirect("/templates", http.StatusFound)
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{})
	})

	controller.Router.Post("/:session/add-area", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))

		if s, ok := controller.sessions[sessionId]; ok {
			var postData rokTemplateArea

			c.BodyParser(&postData)

			if len(strings.TrimSpace(postData.Name)) > 0 {
				if postData.Type == "number" {
					s.schema[postData.Name] = schema.NewNumberField(&schema.OCRCrop{
						X: postData.X,
						Y: postData.Y,
						W: postData.W,
						H: postData.H,
					})
				} else {
					s.schema[postData.Name] = schema.NewTextField(&schema.OCRCrop{
						X: postData.X,
						Y: postData.Y,
						W: postData.W,
						H: postData.H,
					}, "eng")
				}
			}

			return c.Status(http.StatusOK).JSON(fiber.Map{
				"schema":      s.schema,
				"checkpoints": s.checkpoints,
			})
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{})
	})

	controller.Router.Post("/:session/add-checkpoint", func(c *fiber.Ctx) error {
		sessionId := utils.CopyString(c.Params("session"))
		if s, ok := controller.sessions[sessionId]; ok {
			var postData rokCropCoordinates

			c.BodyParser(&postData)

			img, _ := imgutils.ReadImage(s.imagePath)

			cropArea := schema.OCRCrop{
				X: postData.X,
				Y: postData.Y,
				W: postData.W,
				H: postData.H,
			}

			sub, _ := imgutils.CropImage(img, cropArea.CropRectange())
			hash, _ := goimagehash.DifferenceHash(sub)

			s.checkpoints = append(s.checkpoints, schema.OCRCheckpoint{
				Fingerprint: fmt.Sprintf("%x", hash.GetHash()),
				Crop:        &cropArea,
			})

			controller.sessions[sessionId] = s

			return c.JSON(fiber.Map{
				"schema":      s.schema,
				"checkpoints": s.checkpoints,
			})
		}

		return c.Status(http.StatusNotFound).JSON(fiber.Map{})
	})
}
