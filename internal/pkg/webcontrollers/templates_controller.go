package webcontrollers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rokmonster/ocr/internal/pkg/rokocr/tesseractutils"
	imgutils2 "github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/stringutils"

	"github.com/gin-gonic/gin/binding"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
	log "github.com/sirupsen/logrus"
)

type TemplateMakerSession struct {
	imagePath   string
	schema      map[string]schema.OCRSchema
	checkpoints []schema.OCRCheckpoint
}

type TemplatesController struct {
	sessions     map[string]TemplateMakerSession
	templatesDir string
	tessdataDir  string
}

func NewTemplatesController(templateDir, tessdataDir string) *TemplatesController {
	return &TemplatesController{
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

func (controller *TemplatesController) makeTable(s map[string]schema.OCRSchema) []schema.OCRTableField {
	var result []schema.OCRTableField

	for k := range s {
		result = append(result, schema.OCRTableField{
			Title: k,
			Field: k,
			Bold:  false,
			Color: "",
		})
	}

	return result
}

func (controller *TemplatesController) buildTemplate(id string, s TemplateMakerSession) (*schema.OCRTemplate, error) {
	img, err := imgutils2.ReadImageFile(s.imagePath)
	if err != nil {
		return nil, err
	}

	hash, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return nil, err
	}

	return &schema.OCRTemplate{
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
	}, nil
}

func (controller *TemplatesController) ListTemplates(c *gin.Context) {
	c.HTML(http.StatusOK, "templates.html", gin.H{
		"userdata":  c.MustGet(AuthUserData),
		"templates": schema.LoadTemplates(controller.templatesDir),
	})
}

func (controller *TemplatesController) NewTemplateForm(c *gin.Context) {
	c.HTML(http.StatusOK, "templatemaker_upload.html", gin.H{
		"userdata": c.MustGet(AuthUserData),
	})
}

func (controller *TemplatesController) NewTemplatePost(c *gin.Context) {
	// create session id
	sessionId := stringutils.Random(12)
	// handle file upload...
	file, _ := c.FormFile("image")
	dst := os.TempDir() + "/" + sessionId + filepath.Ext(file.Filename)
	_ = c.SaveUploadedFile(file, dst)

	log.Debugf("Uploaded file: %s", dst)

	controller.sessions[sessionId] = TemplateMakerSession{
		imagePath:   dst,
		checkpoints: []schema.OCRCheckpoint{},
		schema:      make(map[string]schema.OCRSchema),
	}

	c.Redirect(http.StatusFound, "/templates/"+sessionId)
}

func (controller *TemplatesController) EditTemplateByID(c *gin.Context) {
	if _, ok := controller.sessions[c.Param("session")]; ok {
		// check if session exists;
		c.HTML(http.StatusOK, "templatemaker.html", gin.H{
			"userdata":  c.MustGet(AuthUserData),
			"sessionId": c.Param("session"),
		})
		return
	}
	c.Redirect(http.StatusFound, "/templates")
}

func (controller *TemplatesController) GetTemplateImage(c *gin.Context) {
	imagePath := controller.sessions[c.Param("session")].imagePath
	c.File(imagePath)
}

func (controller *TemplatesController) TestTemplateByID(c *gin.Context) {
	if s, ok := controller.sessions[c.Param("session")]; ok {
		img, err := imgutils2.ReadImageFile(s.imagePath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
			return
		}

		template, err := controller.buildTemplate(c.Param("session"), s)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"fingerprint": fmt.Sprintf("%x", template.Hash().GetHash()),
			"results":     tesseractutils.ParseImage("test", img, *template, os.TempDir(), controller.tessdataDir).Data,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{})
}

func (controller *TemplatesController) ExportTemplateByID(c *gin.Context) {
	if s, ok := controller.sessions[c.Param("session")]; ok {

		template, err := controller.buildTemplate(time.Now().Format("20060102_150405"), s)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"err": err,
			})
			c.Abort()
			return
		}

		bytes, _ := json.MarshalIndent(template, "", "  ")
		_ = os.WriteFile(fmt.Sprintf("%s/builder_%s.json", controller.templatesDir, c.Param("session")), bytes, 0644)
		c.Redirect(http.StatusFound, "/templates")
		return
	}

	c.JSON(http.StatusNotFound, gin.H{})
}

func (controller *TemplatesController) AddAreaOnTemplate(c *gin.Context) {
	if s, ok := controller.sessions[c.Param("session")]; ok {
		var postData rokTemplateArea

		_ = c.MustBindWith(&postData, binding.JSON)

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
}

func (controller *TemplatesController) AddCheckpointOnTemplate(c *gin.Context) {
	sessionId := c.Param("session")
	if s, ok := controller.sessions[sessionId]; ok {
		var postData rokCropCoordinates

		_ = c.MustBindWith(&postData, binding.JSON)

		img, _ := imgutils2.ReadImageFile(s.imagePath)

		cropArea := schema.OCRCrop{
			X: postData.X,
			Y: postData.Y,
			W: postData.W,
			H: postData.H,
		}

		sub, _ := imgutils2.CropImage(img, cropArea.CropRectangle())
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
}
