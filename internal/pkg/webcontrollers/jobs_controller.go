package webcontrollers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/fileutils"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/stringutils"
	bolt "go.etcd.io/bbolt"
)

type JobsController struct {
	Router *gin.RouterGroup
	db     *bolt.DB
}

func NewJobsController(router *gin.RouterGroup, db *bolt.DB) *JobsController {
	return &JobsController{
		Router: router,
		db:     db,
	}
}

type OCRJob struct {
	ID      uint64                  `json:"id"`
	Name    string                  `json:"name"`
	Results []ocrschema.OCRResponse `json:"results,omitempty"`
	Status  string                  `json:"status,omitempty"`
}

func (job *OCRJob) MediaDirectory() string {
	return fmt.Sprintf("./media/job_%v", job.ID)
}

func (controller *JobsController) getJobs() []OCRJob {
	jobs := []OCRJob{}

	_ = controller.db.View(func(t *bolt.Tx) error {
		if j := t.Bucket([]byte("jobs")); j != nil {
			j.ForEach(func(k, v []byte) error {
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
			return nil
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

func (controller *JobsController) updateJobResults(id uint64, results []ocrschema.OCRResponse) error {
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

func (controller *JobsController) createJob(jobName string) {
	_ = controller.db.Update(func(t *bolt.Tx) error {
		bucket, _ := t.CreateBucketIfNotExists([]byte("jobs"))

		id, _ := bucket.NextSequence()

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
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func (controller *JobsController) Setup() {
	// List all the jobs
	controller.Router.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, "jobs.html", gin.H{
			"jobs": controller.getJobs(),
		})
	})

	controller.Router.GET("/create", func(c *gin.Context) {
		controller.createJob(fmt.Sprintf("Job: %v", time.Now().Format("2006-01-02 15:04:05")))
		c.Redirect(http.StatusFound, "/jobs")
	})

	controller.Router.GET("/:id", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		c.HTML(http.StatusOK, "job_edit.html", gin.H{
			"job":   job,
			"files": controller.getJobFiles(id),
		})
	})

	controller.Router.GET("/:id/start", func(c *gin.Context) {
		// TODO: Edit job here
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		go func(job *OCRJob) {
			controller.updateJobStatus(job.ID, "Started")
			log.Debugf("Processing job: %v", job)
			mediaDir := job.MediaDirectory()

			templates := rokocr.LoadTemplates("./templates")
			if len(templates) > 0 {
				log.Debugf("Loaded %v templates", len(templates))
				template := rokocr.FindTemplate(mediaDir, templates)
				data := rokocr.RunRecognition(mediaDir, "./tessdata", template, false)
				controller.updateJobResults(job.ID, data)
				controller.updateJobStatus(job.ID, "Completed")
			} else {
				log.Warnf("No compatible template found")
				controller.updateJobStatus(job.ID, "Failed, no template found")
			}
		}(job)

		c.Redirect(http.StatusFound, "/jobs")
	})

	controller.Router.GET("/:id/results", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		c.HTML(http.StatusOK, "job_results.html", gin.H{
			"job":   job,
			"files": controller.getJobFiles(id),
		})
	})

	controller.Router.POST("/:id/upload", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)

		os.MkdirAll(job.MediaDirectory(), os.ModePerm)

		// move uploaded file
		file, _ := c.FormFile("file")
		dst := fmt.Sprintf("%s/%s", job.MediaDirectory(), stringutils.Random(8))
		c.SaveUploadedFile(file, dst)

		c.JSON(http.StatusOK, gin.H{
			"destination": dst,
		})
	})

	controller.Router.GET("/:id/delete", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		controller.deleteJob(id)
		c.Redirect(http.StatusFound, "/jobs")
	})

}
