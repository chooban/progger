package db

import (
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func createDb() *gorm.DB {
	return Init("file::memory:?cache=shared")
}

func TestSaveIssue(t *testing.T) {
	db := createDb()

	issue := Issue{
		Publication: Publication{Title: "Test Publication"},
		IssueNumber: 123,
		Episodes: []Episode{
			{
				Title:  "Test Episode",
				Part:   1,
				Series: Series{Title: "Test Series"},
			},
		},
	}
	SaveIssues(db, []Issue{issue})

	var count int64
	db.Model(Publication{}).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(Issue{}).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(Series{}).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(Episode{}).Count(&count)
	assert.Equal(t, int64(1), count)
}
