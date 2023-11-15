package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type Publication struct {
	gorm.Model
	Title string `gorm:"CHECK:title <> NULL;uniqueIndex"`
}

type Series struct {
	gorm.Model
	Title string `gorm:"CHECK:title <> NULL;CHECK:title <> \"\";uniqueIndex"`
}

type Issue struct {
	gorm.Model
	PublicationID uint `gorm:"uniqueIndex:idx_pub_issue"`
	Publication   Publication
	IssueNumber   int `gorm:"CHECK:issue_number >= 0;uniqueIndex:idx_pub_issue"`
	Episodes      []Episode
	Filename      string
}

type Episode struct {
	gorm.Model
	Title    string
	Part     int `gorm:"CHECK:part >= 0"`
	IssueID  uint
	Issue    Issue
	SeriesID uint
	Series   Series
	PageFrom int
	PageThru int
	Script   []*Creator `gorm:"many2many:episode_writers"`
	Art      []*Creator `gorm:"many2many:episode_artists"`
	Colours  []*Creator `gorm:"many2many:episode_colourists"`
	Letters  []*Creator `gorm:"many2many:episode_letterers"`
}

type Creator struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
}

type EpisodeWriter struct {
	EpisodeID uint `gorm:"primaryKey"`
	CreatorID uint `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type EpisodeArtist struct {
	EpisodeID uint `gorm:"primaryKey"`
	CreatorID uint `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type EpisodeColourist struct {
	EpisodeID uint `gorm:"primaryKey"`
	CreatorID uint `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}
type EpisodeLetterer struct {
	EpisodeID uint `gorm:"primaryKey"`
	CreatorID uint `gorm:"primaryKey"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func Init(dbName string) *gorm.DB {
	gormdb, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = gormdb.AutoMigrate(
		&Issue{},
		&Episode{},
		&Series{},
		&Publication{},
		&Creator{},
		&EpisodeWriter{},
		&EpisodeArtist{},
		&EpisodeColourist{},
		&EpisodeLetterer{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := gormdb.SetupJoinTable(&Episode{}, "Script", &EpisodeWriter{}); err != nil {
		log.Fatalf("Failed to setup join table: %v", err)
	}

	return gormdb
}

func SaveIssues(db *gorm.DB, issues []Issue) {
	for _, issue := range issues {
		SaveIssue(db, issue)
	}
}

func SaveIssue(db *gorm.DB, issue Issue) {
	db.FirstOrCreate(
		&issue.Publication,
		Publication{Title: issue.Publication.Title},
	)
	issue.PublicationID = issue.Publication.ID
	res := db.
		Where(&Issue{PublicationID: issue.PublicationID}).
		Attrs(&Issue{IssueNumber: issue.IssueNumber}).
		Omit(clause.Associations).
		FirstOrCreate(&issue, &Issue{
			PublicationID: issue.PublicationID,
			IssueNumber:   issue.IssueNumber,
			Filename:      issue.Filename,
		})

	if res.RowsAffected == 0 {
		return
	}

	for _, e := range issue.Episodes {
		for _, w := range e.Script {
			db.Where(&Creator{Name: w.Name}).FirstOrCreate(&w, Creator{Name: w.Name})
		}
		db.Where(&Series{Title: e.Series.Title}).FirstOrCreate(&e.Series, Series{Title: e.Series.Title})
		e.SeriesID = e.Series.ID
		e.IssueID = issue.ID
		db.Create(&e)

		if err := db.Model(&e).Association("Script").Append(&e.Script); err != nil {
			fmt.Println(err)
		}
	}
}
