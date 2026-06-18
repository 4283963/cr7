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

type Script struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	CharacterID string  `json:"characterId"`
	Duration    float64 `json:"duration"`
	Tracks      []Track `json:"tracks"`
	FPS         int     `json:"fps"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type ScriptListEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CharacterID string `json:"characterId"`
	Duration    float64 `json:"duration"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}
