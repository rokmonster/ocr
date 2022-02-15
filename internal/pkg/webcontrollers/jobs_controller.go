package webcontrollers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	ID   uint64 `json:"id"`
	Name string `json:"name"`
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
		// TODO: Edit job here
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		job := controller.getJob(id)
		log.Printf("edit job: %v", job)
		c.HTML(http.StatusOK, "job_edit.html", gin.H{
			"job": job,
		})
	})

	controller.Router.POST("/:id", func(c *gin.Context) {
		// TODO: Edit job here
		c.Redirect(http.StatusFound, "/jobs")
	})

	controller.Router.GET("/:id/delete", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 0, 64)
		log.Warnf("delete job: %+v", id)
		controller.deleteJob(id)
		c.Redirect(http.StatusFound, "/jobs")
	})

	controller.Router.GET("/:id/start", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/jobs")
	})
}
