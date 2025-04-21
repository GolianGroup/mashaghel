package presenters

import (
	"mashaghel/internal/repositories/models"
	"time"
)

type videoPresenter struct {
	Key         string   `json:"id"`    // Changed from key to id
	Name        string   `json:"title"` // Changed from name to title
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	ViewCount   int      `json:"view_count"`   // Changed from views
	Type        string   `json:"content_type"` // Changed from type
	IsPublic    bool     `json:"is_public"`    // Changed from publishable
	CreatedAt   string   `json:"created_at"`   // New field
}

func NewVideoPresenter(video *models.Video) Presenter {
	return &videoPresenter{
		Key:         video.Key,
		Name:        video.Name,
		Description: video.Description,
		Categories:  video.Categories,
		ViewCount:   video.Views,
		Type:        video.Type,
		IsPublic:    video.Publishable,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func (p *videoPresenter) Present() interface{} {
	return p
}
