package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
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
}

func Init(dbName string) *gorm.DB {
	gormdb, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = gormdb.AutoMigrate(&Issue{}, &Episode{}, &Series{}, &Publication{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
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
	db.Omit(clause.Associations).FirstOrCreate(&issue)
	for _, e := range issue.Episodes {
		db.FirstOrCreate(&e.Series, Series{Title: e.Series.Title})
		e.SeriesID = e.Series.ID
		e.IssueID = issue.ID
		db.Create(&e)
	}
}
