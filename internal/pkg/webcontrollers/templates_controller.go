package webcontrollers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
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
	Router       *gin.RouterGroup
	sessions     map[string]TemplateMakerSession
	templatesDir string
	tessdataDir  string
}

func NewTemplatesController(router *gin.RouterGroup, templateDir, tessdataDir string) *TemplatesController {
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

	controller.Router.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, "templatemaker_upload.html", gin.H{})
	})

	// create new session, and redirect
	controller.Router.POST("", func(c *gin.Context) {
		// create session id
		sessionId := stringutils.Random(12)
		// handle file upload...
		file, _ := c.FormFile("image")
		dst := os.TempDir() + "/" + sessionId + filepath.Ext(file.Filename)
		c.SaveUploadedFile(file, dst)

		logrus.Debugf("Uploaded file: %s", dst)

		controller.sessions[sessionId] = TemplateMakerSession{
			imagePath:   dst,
			checkpoints: []schema.OCRCheckpoint{},
			schema:      make(map[string]schema.ROKOCRSchema),
		}

		c.Redirect(http.StatusFound, "/templates/"+sessionId)
	})

	controller.Router.GET("/:session", func(c *gin.Context) {
		if _, ok := controller.sessions[c.Param("session")]; ok {
			// check if session exists;
			c.HTML(http.StatusOK, "templatemaker.html", gin.H{
				"sessionId": c.Param("session"),
			})
			return
		}
		c.Redirect(http.StatusFound, "/templates")

	})

	controller.Router.GET("/:session/image", func(c *gin.Context) {
		imagePath := controller.sessions[c.Param("session")].imagePath
		c.File(imagePath)
	})

	controller.Router.POST("/:session/scan", func(c *gin.Context) {
		if s, ok := controller.sessions[c.Param("session")]; ok {
			img, _ := imgutils.ReadImage(s.imagePath)
			template := controller.buildTemplate(c.Param("session"), s)

			c.JSON(http.StatusOK, gin.H{
				"fingerprint": fmt.Sprintf("%x", template.Hash().GetHash()),
				"results":     rokocr.ParseImage("test", img, template, os.TempDir(), "./tessdata"),
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{})
	})

	controller.Router.GET("/:session/export", func(c *gin.Context) {
		if s, ok := controller.sessions[c.Param("session")]; ok {
			template := controller.buildTemplate(c.Param("session"), s)
			bytes, _ := json.MarshalIndent(template, "", "    ")
			os.WriteFile(fmt.Sprintf("%s/builder_%s.json", controller.templatesDir, c.Param("session")), bytes, 0644)
			c.Redirect(http.StatusFound, "/templates")
			return
		}

		c.JSON(http.StatusNotFound, gin.H{})
	})

	controller.Router.POST("/:session/add-area", func(c *gin.Context) {
		if s, ok := controller.sessions[c.Param("session")]; ok {
			var postData rokTemplateArea

			c.BindJSON(&postData)

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

			c.JSON(http.StatusOK, gin.H{
				"schema":      s.schema,
				"checkpoints": s.checkpoints,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{})
	})

	controller.Router.POST("/:session/add-checkpoint", func(c *gin.Context) {
		sessionId := c.Param("session")
		if s, ok := controller.sessions[sessionId]; ok {
			var postData rokCropCoordinates

			c.BindJSON(&postData)

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

			c.JSON(http.StatusOK, gin.H{
				"schema":      s.schema,
				"checkpoints": s.checkpoints,
			})

			return
		}

		c.JSON(http.StatusNotFound, gin.H{})
	})
}
