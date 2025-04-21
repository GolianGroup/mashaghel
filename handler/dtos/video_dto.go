package dto

type Video struct {
	Publishable bool     `json:"publishable" validate:"required"`
	Categories  []string `json:"categories" validate:"required,dive,required"`
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name" validate:"required"`
	Type        string   `json:"type,omitempty" validate:"oneof=movie series tvshow"`
}

type VideoUpdate struct {
	Key         string   `json:"key" validate:"required"`
	Categories  []string `json:"categories" validate:"required,dive,required"`
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name" validate:"required"`
	Views       int      `json:"views,omitempty" validate:"gte=0"`
}
