package async

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type HttpJobHandler struct {
	jm JobManager
}

func (h *HttpJobHandler) Get(c *gin.Context) {

	id, err := uuid.FromString(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	job, err := h.jm.GetByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job": job,
	})
}

type CreateIn struct {
	Name      string                 `json:"name" binding:"required"`
	Functions []*Function            `json:"functions" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

func (h *HttpJobHandler) Create(c *gin.Context) {
	var err error

	var in CreateIn
	if err = c.Bind(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var job *Job
	if job, err = h.jm.Create(in.Name, in.Functions, in.Data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job": job,
	})
}

func NewHttpJobHandler(e *gin.Engine, jm JobManager) {

	handler := &HttpJobHandler{
		jm: jm,
	}

	e.GET("/v1/job/:job_id", handler.Get)
	e.POST("/v1/job", handler.Create)
}
