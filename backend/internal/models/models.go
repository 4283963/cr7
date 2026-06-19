package models

type Bone struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ParentID  string  `json:"parentId"`
	Length    float64 `json:"length"`
	BaseAngle float64 `json:"baseAngle"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	MinAngle  float64 `json:"minAngle"`
	MaxAngle  float64 `json:"maxAngle"`
	Color     string  `json:"color"`
	Width     float64 `json:"width"`
}

type Character struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Bones       []Bone  `json:"bones"`
	RootBoneID  string  `json:"rootBoneId"`
	BaseX       float64 `json:"baseX"`
	BaseY       float64 `json:"baseY"`
	Scale       float64 `json:"scale"`
}

type Keyframe struct {
	Time   float64            `json:"time"`
	Values map[string]float64 `json:"values"`
}

type Track struct {
	BoneID     string      `json:"boneId"`
	BoneName   string      `json:"boneName"`
	Keyframes  []Keyframe  `json:"keyframes"`
	InterpType string      `json:"interpType"`
}

type BeatPoint struct {
	Time    float64 `json:"time"`
	Label   string  `json:"label,omitempty"`
	Enabled bool    `json:"enabled"`
}

type AudioTrack struct {
	FileName   string      `json:"fileName"`
	FileSize   int64       `json:"fileSize"`
	Duration   float64     `json:"duration"`
	Volume     float64     `json:"volume"`
	StartTime  float64     `json:"startTime"`
	Beats      []BeatPoint `json:"beats"`
	SyncTolerance float64  `json:"syncTolerance"`
}

type Script struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CharacterID string     `json:"characterId"`
	Duration    float64    `json:"duration"`
	Tracks      []Track    `json:"tracks"`
	FPS         int        `json:"fps"`
	AudioTrack  *AudioTrack `json:"audioTrack,omitempty"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
}

type ScriptListEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CharacterID string `json:"characterId"`
	Duration    float64 `json:"duration"`
	HasAudio    bool    `json:"hasAudio"`
	BeatCount   int     `json:"beatCount"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

