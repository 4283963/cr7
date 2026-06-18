package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"shadow-puppet-backend/internal/models"
	"shadow-puppet-backend/internal/storage"
)

type Handler struct {
	store *storage.Storage
}

func New(store *storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) ListCharacters(c *gin.Context) {
	characters, err := h.store.ListCharacters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, characters)
}

func (h *Handler) GetCharacter(c *gin.Context) {
	id := c.Param("id")
	character, err := h.store.GetCharacter(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Character not found"})
		return
	}
	c.JSON(http.StatusOK, character)
}

func (h *Handler) CreateCharacter(c *gin.Context) {
	var character models.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if character.ID == "" {
		character.ID = fmt.Sprintf("char_%d", time.Now().Unix())
	}
	if character.Scale == 0 {
		character.Scale = 1.0
	}
	if err := h.store.SaveCharacter(&character); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, character)
}

func (h *Handler) UpdateCharacter(c *gin.Context) {
	id := c.Param("id")
	var character models.Character
	if err := c.ShouldBindJSON(&character); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	character.ID = id
	if err := h.store.SaveCharacter(&character); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, character)
}

func (h *Handler) DeleteCharacter(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteCharacter(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Character deleted"})
}

func (h *Handler) ListScripts(c *gin.Context) {
	scripts, err := h.store.ListScripts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scripts)
}

func (h *Handler) GetScript(c *gin.Context) {
	id := c.Param("id")
	script, err := h.store.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}
	c.JSON(http.StatusOK, script)
}

func (h *Handler) CreateScript(c *gin.Context) {
	var script models.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if script.ID == "" {
		script.ID = fmt.Sprintf("script_%d", time.Now().Unix())
	}
	if script.FPS == 0 {
		script.FPS = 30
	}
	if err := h.store.SaveScript(&script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, script)
}

func (h *Handler) UpdateScript(c *gin.Context) {
	id := c.Param("id")
	var script models.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	script.ID = id
	if err := h.store.SaveScript(&script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, script)
}

func (h *Handler) DeleteScript(c *gin.Context) {
	id := c.Param("id")
	if err := h.store.DeleteScript(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Script deleted"})
}

func (h *Handler) DuplicateScript(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "Copy"
	}
	newScript, err := h.store.DuplicateScript(id, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newScript)
}
