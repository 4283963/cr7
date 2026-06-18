package main

import (
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"shadow-puppet-backend/internal/handlers"
	"shadow-puppet-backend/internal/storage"
)

func main() {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(filepath.Dir(filepath.Dir(b)))

	charactersDir := filepath.Join(basepath, "data", "characters")
	scriptsDir := filepath.Join(basepath, "data", "scripts")

	store := storage.New(charactersDir, scriptsDir)
	h := handlers.New(store)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		characters := api.Group("/characters")
		{
			characters.GET("", h.ListCharacters)
			characters.GET("/:id", h.GetCharacter)
			characters.POST("", h.CreateCharacter)
			characters.PUT("/:id", h.UpdateCharacter)
			characters.DELETE("/:id", h.DeleteCharacter)
		}

		scripts := api.Group("/scripts")
		{
			scripts.GET("", h.ListScripts)
			scripts.GET("/:id", h.GetScript)
			scripts.POST("", h.CreateScript)
			scripts.PUT("/:id", h.UpdateScript)
			scripts.DELETE("/:id", h.DeleteScript)
			scripts.POST("/:id/duplicate", h.DuplicateScript)
		}
	}

	log.Println("Shadow Puppet Backend Server starting on :8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
