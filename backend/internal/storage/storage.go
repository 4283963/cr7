package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"shadow-puppet-backend/internal/models"
)

type Storage struct {
	charactersDir string
	scriptsDir    string
}

func New(charactersDir, scriptsDir string) *Storage {
	os.MkdirAll(charactersDir, 0755)
	os.MkdirAll(scriptsDir, 0755)
	return &Storage{
		charactersDir: charactersDir,
		scriptsDir:    scriptsDir,
	}
}

func (s *Storage) characterPath(id string) string {
	return filepath.Join(s.charactersDir, id+".json")
}

func (s *Storage) scriptPath(id string) string {
	return filepath.Join(s.scriptsDir, id+".json")
}

func (s *Storage) ListCharacters() ([]models.Character, error) {
	files, err := os.ReadDir(s.charactersDir)
	if err != nil {
		return nil, err
	}

	var characters []models.Character
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.charactersDir, f.Name()))
		if err != nil {
			continue
		}
		var c models.Character
		if err := json.Unmarshal(data, &c); err != nil {
			continue
		}
		characters = append(characters, c)
	}
	return characters, nil
}

func (s *Storage) GetCharacter(id string) (*models.Character, error) {
	data, err := os.ReadFile(s.characterPath(id))
	if err != nil {
		return nil, err
	}
	var c models.Character
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Storage) SaveCharacter(c *models.Character) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.characterPath(c.ID), data, 0644)
}

func (s *Storage) DeleteCharacter(id string) error {
	return os.Remove(s.characterPath(id))
}

func (s *Storage) ListScripts() ([]models.ScriptListEntry, error) {
	files, err := os.ReadDir(s.scriptsDir)
	if err != nil {
		return nil, err
	}

	var scripts []models.ScriptListEntry
	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.scriptsDir, f.Name()))
		if err != nil {
			continue
		}
		var s models.Script
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		scripts = append(scripts, models.ScriptListEntry{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			CharacterID: s.CharacterID,
			Duration:    s.Duration,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].UpdatedAt > scripts[j].UpdatedAt
	})

	return scripts, nil
}

func (s *Storage) GetScript(id string) (*models.Script, error) {
	data, err := os.ReadFile(s.scriptPath(id))
	if err != nil {
		return nil, err
	}
	var script models.Script
	if err := json.Unmarshal(data, &script); err != nil {
		return nil, err
	}
	return &script, nil
}

func (s *Storage) SaveScript(script *models.Script) error {
	now := time.Now().Format(time.RFC3339)
	if script.CreatedAt == "" {
		script.CreatedAt = now
	}
	script.UpdatedAt = now

	if script.FPS == 0 {
		script.FPS = 30
	}

	data, err := json.MarshalIndent(script, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.scriptPath(script.ID), data, 0644)
}

func (s *Storage) DeleteScript(id string) error {
	return os.Remove(s.scriptPath(id))
}

func (s *Storage) DuplicateScript(id, newName string) (*models.Script, error) {
	script, err := s.GetScript(id)
	if err != nil {
		return nil, err
	}

	newScript := *script
	newScript.ID = fmt.Sprintf("script_%d", time.Now().Unix())
	newScript.Name = newName
	now := time.Now().Format(time.RFC3339)
	newScript.CreatedAt = now
	newScript.UpdatedAt = now

	if err := s.SaveScript(&newScript); err != nil {
		return nil, err
	}

	return &newScript, nil
}
