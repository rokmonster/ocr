package webcontrollers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/rokmonster/ocr/internal/pkg/rokocr/tesseractutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rokmonster/ocr/internal/pkg/ocrschema"
	"github.com/rokmonster/ocr/internal/pkg/rokocr"
	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type JobsController struct {
	db *bolt.DB
}

func NewJobsController(db *bolt.DB) *JobsController {
	return &JobsController{
		db: db,
	}
}

type OCRJob struct {
	ID       uint64                `json:"id"`
	Name     string                `json:"name"`
	Results  []ocrschema.OCRResult `json:"results,omitempty"`
	Status   string                `json:"status,omitempty"`
	Template ocrschema.OCRTemplate `json:"template,omitempty"`
}

func (job *OCRJob) MediaDirectory() string {
	return fmt.Sprintf("./media/job_%v", job.ID)
}

func (controller *JobsController) getJobs() []OCRJob {
	var jobs []OCRJob

	_ = controller.db.View(func(t *bolt.Tx) error {
		if j := t.Bucket([]byte("jobs")); j != nil {
			return j.ForEach(func(k, v []byte) error {
				var job OCRJob
				if err := json.Unmarshal(v, &job); err != nil {
					return err
				}

				if len(strings.TrimSpace(job.Status)) == 0 {
					job.Status = fmt.Sprintf("Pending: 0/%v processed", len(controller.getJobFiles(job.ID)))
				}

				jobs = append(jobs, job)
				return nil
			})
		}
		return fmt.Errorf("bucket not found")
	})

	return jobs
}

func (controller *JobsController) deleteJob(id uint64) {
	_ = controller.db.Update(func(t *bolt.Tx) error {
		return t.Bucket([]byte("jobs")).Delete(itob(id))
	})
}

func (controller *JobsController) getJobFiles(id uint64) []string {
	job := controller.getJob(id)
	return fileutils.GetFilesInDirectory(job.MediaDirectory())
}

func (controller *JobsController) updateJob(id uint64, fn func(*OCRJob) *OCRJob) error {
	return controller.db.Update(func(t *bolt.Tx) error {
		var job *OCRJob

		bucket := t.Bucket([]byte("jobs"))

		bytes := bucket.Get(itob(id))
		err := json.Unmarshal(bytes, &job)
		if err != nil {
			return err
		}

		job = fn(job)

		buf, err := json.Marshal(job)
		if err != nil {
			return err
		}

		return bucket.Put(itob(job.ID), buf)
	})
}

func (controller *JobsController) updateJobStatus(id uint64, status string) error {
	return controller.updateJob(id, func(job *OCRJob) *OCRJob {
		job.Status = status
		return job
	})
}

func (controller *JobsController) updateJobTemplate(id uint64, template ocrschema.OCRTemplate) error {
	return controller.updateJob(id, func(job *OCRJob) *OCRJob {
		job.Template = template
		return job
	})
}

func (controller *JobsController) updateJobResults(id uint64, results []ocrschema.OCRResult) error {
	return controller.updateJob(id, func(job *OCRJob) *OCRJob {
		job.Results = results
		return job
	})
}

func (controller *JobsController) getJob(id uint64) *OCRJob {
	var job *OCRJob

	_ = controller.db.View(func(t *bolt.Tx) error {
		bytes := t.Bucket([]byte("jobs")).Get(itob(id))
		return json.Unmarshal(bytes, &job)
	})

	return job
}

func (controller *JobsController) createJob(jobName string) (uint64, error) {
	id := uint64(0)

	err := controller.db.Update(func(t *bolt.Tx) error {
		bucket, _ := t.CreateBucketIfNotExists([]byte("jobs"))

		id, _ = bucket.NextSequence()

		u := OCRJob{
			Name: jobName,
			ID:   id,
		}

		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}

		return bucket.Put(itob(u.ID), buf)
	})

	return id, err
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func (controller *JobsController) Setup(router *gin.RouterGroup) {
	// List all the jobs
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "jobs.html", gin.H{
			"jobs": controller.getJobs(),
		})
	})

	router.GET("/create", func(c *gin.Context) {
		id, err := controller.createJob(fmt.Sprintf("Job: %v", time.Now().Format("2006-01-02 15:04:05")))
		if err == nil {
			c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/%v", id))
		} else {
			c.Redirect(http.StatusFound, "/jobs")
		}
	})

	router.GET("/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		c.HTML(http.StatusOK, "job_edit.html", gin.H{
			"job":   job,
			"files": controller.getJobFiles(id),
		})
	})

	router.GET("/:id/start", func(c *gin.Context) {
		// TODO: Edit job here
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		go func(job *OCRJob) {
			log.Debugf("Processing job: %v", job)

			index := 1
			fileCount := len(controller.getJobFiles(job.ID))

			// clean results & update status
			_ = controller.updateJobResults(job.ID, []ocrschema.OCRResult{})
			_ = controller.updateJobStatus(job.ID, fmt.Sprintf("Processing: %v/%v", index, fileCount))

			mediaDir := job.MediaDirectory()

			templates := ocrschema.LoadTemplates("./templates")
			if len(templates) > 0 {
				log.Debugf("Loaded %v templates", len(templates))
				template := ocrschema.FindTemplate(mediaDir, templates)
				_ = controller.updateJobTemplate(job.ID, template)

				var data []ocrschema.OCRResult
				for elem := range tesseractutils.RunRecognitionChan(mediaDir, "./tessdata", template, false) {
					data = append(data, elem)
					index = index + 1
					_ = controller.updateJobStatus(job.ID, fmt.Sprintf("Processing: %v/%v", index, fileCount))
					_ = controller.updateJobResults(job.ID, data)
				}

				_ = controller.updateJobResults(job.ID, data)
				_ = controller.updateJobStatus(job.ID, fmt.Sprintf("Completed: %v files", len(data)))
			} else {
				log.Warnf("No compatible template found")
				_ = controller.updateJobStatus(job.ID, "Failed, no template found")
			}
		}(job)

		c.Redirect(http.StatusFound, fmt.Sprintf("/jobs/%v/results", id))
	})

	router.GET("/:id/csv", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		b := new(bytes.Buffer)
		rokocr.WriteCSV(job.Results, job.Template, b)

		c.Data(http.StatusOK, "text/plain", b.Bytes())
	})

	router.GET("/:id/results", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		c.HTML(http.StatusOK, "job_results.html", gin.H{
			"job":   job,
			"files": controller.getJobFiles(id),
		})
	})

	router.POST("/:id/upload", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		_ = os.MkdirAll(job.MediaDirectory(), os.ModePerm)

		// move uploaded file
		file, _ := c.FormFile("file")
		dst := filepath.Join(job.MediaDirectory(), filepath.Base(file.Filename))
		_ = c.SaveUploadedFile(file, dst)

		c.JSON(http.StatusOK, gin.H{
			"destination": dst,
		})
	})

	router.GET("/:id/delete", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		controller.deleteJob(id)
		c.Redirect(http.StatusFound, "/jobs")
	})

}
