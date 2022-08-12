package www

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rokmonster/ocr/internal/pkg/rokocr/opencvutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/sirupsen/logrus"
)

type APIController struct {
	tessdataDir string
}

func NewAPIController(tessDir string) *APIController {
	return &APIController{tessdataDir: tessDir}
}

func (controller *APIController) ScanHOH(c *gin.Context) {
	var request struct {
		ImageURL string `json:"url"`
	}
	err := c.MustBindWith(&request, binding.JSON)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": fmt.Errorf("error while parsing request url"),
		})
		return
	}

	logrus.Infof("Got request to parse HOH from: %v", request.ImageURL)

	response, err := http.Get(request.ImageURL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{})
		return
	}

	defer response.Body.Close()
	img, _ := imgutils.ReadImage(response.Body)

	c.JSON(http.StatusOK, opencvutils.HOHScan(img, controller.tessdataDir))
}
