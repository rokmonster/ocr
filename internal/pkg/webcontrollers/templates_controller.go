package webcontrollers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/stringutils"
)

type TemplateMakerSession struct {
	imagePath string
	schema    map[string]schema.ROKOCRSchema
}

type TemplatesController struct {
	Router   *gin.RouterGroup
	sessions map[string]TemplateMakerSession
}

func NewTemplatesController(router *gin.RouterGroup) *TemplatesController {
	return &TemplatesController{
		Router:   router,
		sessions: make(map[string]TemplateMakerSession),
	}
}

// Binding from JSON
type rokTemplateArea struct {
	Name string `form:"name" json:"name" xml:"name"  binding:"required"`
	X    int    `form:"x" json:"x,string" xml:"x"  binding:"required"`
	Y    int    `form:"y" json:"y,string" xml:"y" binding:"required"`
	W    int    `form:"w" json:"w,string" xml:"w" binding:"required"`
	H    int    `form:"h" json:"h,string" xml:"h" binding:"required"`
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
		dst := os.TempDir() + sessionId + filepath.Ext(file.Filename)
		c.SaveUploadedFile(file, dst)

		logrus.Infof("Uploaded file: %s", dst)

		controller.sessions[sessionId] = TemplateMakerSession{
			imagePath: dst,
			schema:    make(map[string]schema.ROKOCRSchema),
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
			hash, _ := goimagehash.DifferenceHash(img)
			c.JSON(http.StatusOK, gin.H{
				"fingerprint": fmt.Sprintf("%x", hash.GetHash()),
				"results": rokocr.ParseImage("test", img, &schema.RokOCRTemplate{
					Title:       fmt.Sprintf("Pending template [%s]", c.Param("session")),
					Version:     "1",
					Fingerprint: fmt.Sprintf("%x", hash.GetHash()),
					Width:       img.Bounds().Dx(),
					Height:      img.Bounds().Dy(),
					Author:      "ROK OCR Template Maker",
					Threshold:   1,
					OCRSchema:   s.schema,
				}, os.TempDir(), "./tessdata"),
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{})

	})

	controller.Router.POST("/:session/add-area", func(c *gin.Context) {
		if s, ok := controller.sessions[c.Param("session")]; ok {
			var postData rokTemplateArea

			c.BindJSON(&postData)

			s.schema[postData.Name] = schema.ROKOCRSchema{
				Languages: []string{"eng"},
				AllowList: []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
				PSM:       7,
				Crop: schema.OCRCrop{
					X: postData.X,
					Y: postData.Y,
					W: postData.W,
					H: postData.H,
				},
			}

			c.JSON(http.StatusOK, gin.H{
				"schema": s.schema,
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{})
	})
}
