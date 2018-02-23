package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type HttpCallbackHandler struct {
	cm CallbackManager
	jm JobManager
}

func (h *HttpCallbackHandler) getCallback(c *gin.Context) (*Callback, error) {
	id, err := uuid.FromString(c.Param("callback_id"))
	if err != nil {
		return nil, err
	}

	callback, err := h.cm.GetByID(id)
	if err != nil {
		return nil, err
	}

	return callback, nil
}

func (h *HttpCallbackHandler) Get(c *gin.Context) {

	callback, err := h.getCallback(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"callback": callback,
	})
}

type CallbackIn struct {
	StatusCode int `json:"status_code" binding:"required"`
}

func (h *HttpCallbackHandler) Callback(c *gin.Context) {

	callback, err := h.getCallback(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var in CallbackIn
	if err = c.Bind(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if callback.IsExpired() {
		logger.Printf("http_callback_handler: callback %s is expired, ignoring", callback.ID)
		c.Status(http.StatusOK)
		return
	}

	if err := h.jm.HandleCallback(callback, in.StatusCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.cm.Delete(callback.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

func NewHttpCallbackHandler(e *gin.Engine, cm CallbackManager, jm JobManager) {

	handler := &HttpCallbackHandler{
		cm: cm,
		jm: jm,
	}

	e.GET("/v1/callback/:callback_id", handler.Get)
	e.POST("/v1/callback/:callback_id", handler.Callback)
}
