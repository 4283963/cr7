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

func (h *Handler) UploadAudio(c *gin.Context) {
	id := c.Param("id")

	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio file provided"})
		return
	}
	defer file.Close()

	ext := ""
	switch header.Header.Get("Content-Type") {
	case "audio/mpeg", "audio/mp3":
		ext = ".mp3"
	case "audio/wav", "audio/x-wav":
		ext = ".wav"
	case "audio/ogg":
		ext = ".ogg"
	case "audio/m4a", "audio/mp4":
		ext = ".m4a"
	}
	fileName := fmt.Sprintf("audio%s", ext)

	storedName, size, err := h.store.SaveAudio(id, fileName, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	script, err := h.store.GetScript(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if script.AudioTrack == nil {
		script.AudioTrack = &models.AudioTrack{}
	}
	script.AudioTrack.FileName = storedName
	script.AudioTrack.FileSize = size
	script.AudioTrack.Volume = 1.0
	script.AudioTrack.StartTime = 0
	if script.AudioTrack.SyncTolerance == 0 {
		script.AudioTrack.SyncTolerance = 0.08
	}
	if script.AudioTrack.Beats == nil {
		script.AudioTrack.Beats = []models.BeatPoint{}
	}

	if err := h.store.SaveScript(script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fileName": storedName,
		"fileSize": size,
		"url":      fmt.Sprintf("/api/scripts/%s/audio/%s", id, storedName),
	})
}

func (h *Handler) GetAudio(c *gin.Context) {
	id := c.Param("id")
	fileName := c.Param("fileName")

	filePath := h.store.AudioPath(id, fileName)
	c.File(filePath)
}

func (h *Handler) DeleteAudio(c *gin.Context) {
	id := c.Param("id")

	script, err := h.store.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}

	if script.AudioTrack != nil && script.AudioTrack.FileName != "" {
		if err := h.store.DeleteAudio(id, script.AudioTrack.FileName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		script.AudioTrack = nil
		if err := h.store.SaveScript(script); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audio deleted"})
}

func (h *Handler) UpdateBeats(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Beats         []models.BeatPoint `json:"beats"`
		SyncTolerance float64            `json:"syncTolerance"`
		StartTime     float64            `json:"startTime"`
		Volume        float64            `json:"volume"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	script, err := h.store.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}

	if script.AudioTrack == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio track exists"})
		return
	}

	script.AudioTrack.Beats = req.Beats
	if req.SyncTolerance > 0 {
		script.AudioTrack.SyncTolerance = req.SyncTolerance
	}
	if req.StartTime >= 0 {
		script.AudioTrack.StartTime = req.StartTime
	}
	if req.Volume >= 0 {
		script.AudioTrack.Volume = req.Volume
	}

	if err := h.store.SaveScript(script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, script)
}

func (h *Handler) AnalyzeBeats(c *gin.Context) {
	id := c.Param("id")

	script, err := h.store.GetScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Script not found"})
		return
	}

	if script.AudioTrack == nil || script.AudioTrack.FileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No audio uploaded"})
		return
	}

	var tempo struct {
		BPM    float64 `json:"bpm"`
		Offset float64 `json:"offset"`
	}
	if err := c.ShouldBindJSON(&tempo); err != nil || tempo.BPM <= 0 {
		tempo.BPM = 100
		tempo.Offset = 0
	}

	beatInterval := 60.0 / tempo.BPM
	duration := script.Duration
	beats := []models.BeatPoint{}

	for t := tempo.Offset; t <= duration+beatInterval; t += beatInterval {
		beats = append(beats, models.BeatPoint{
			Time:    t,
			Label:   "",
			Enabled: true,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"beats":    beats,
		"interval": beatInterval,
		"bpm":      tempo.BPM,
		"offset":   tempo.Offset,
	})
}
