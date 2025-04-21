package presenters

import (
	"mashaghel/internal/repositories/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVideoPresenter(t *testing.T) {
	video := &models.Video{
		Key:         "123",
		Name:        "Test Video",
		Description: "Test Description",
		Categories:  []string{"test"},
		Views:       100,
		Type:        "movie",
		Publishable: true,
	}

	presenter := NewVideoPresenter(video)
	result := presenter.Present().(*videoPresenter)

	assert.Equal(t, video.Key, result.Key)
	assert.Equal(t, video.Name, result.Name)
	assert.Equal(t, video.Views, result.ViewCount)
	assert.Equal(t, video.Publishable, result.IsPublic)
}
