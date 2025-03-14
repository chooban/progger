package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chooban/progger/exporter/api"
	"github.com/sdomino/scribble"
)

var defaultSkipTitles = []string{
	"Interrogation",
	"New Books",
	"Obituary",
	"Tribute",
	"Untitled",
	"Encyclopedia",
}
var defaultKnownTitles = []string{
	"Anderson, Psi-Division",
	"Chimpsky's Law",
	"Counterfeit Girl",
	"Feral & Foe",
	"Lowborn High",
	"Scarlet Traces",
	"Strontium Dog",
	"Strontium Dug",
	"The Fall of Deadworld",
}

type Storage struct {
	storageDir string
	db         *scribble.Driver
}

func (s *Storage) SaveIssues(progs []api.Downloadable) error {
	if s.db == nil {
		return errors.New("db not initialized")
	}
	var err error
	for _, p := range progs {
		err = s.db.Write("proglist", fmt.Sprintf("prog_%d", p.Comic.IssueNumber), p)
		if err != nil {
			break
		}
	}
	return err
}

func (s *Storage) ReadIssues() []api.Downloadable {
	records, err := s.db.ReadAll("proglist")
	if err != nil {
		println(err.Error())
		return make([]api.Downloadable, 0)
	}
	progs := make([]api.Downloadable, 0, len(records))
	for _, p := range records {
		readProg := api.Downloadable{}
		if err := json.Unmarshal(p, &readProg); err != nil {
			fmt.Println("Error", err)
		}
		readProg.Downloaded = false
		progs = append(progs, readProg)
	}
	return progs
}

func (s *Storage) StoreStories(stories []api.Story) error {
	if s.db == nil {
		return errors.New("db not initialized")
	}
	var err error
	for _, p := range stories {
		err = s.db.Write("stories_list", fmt.Sprintf("story_%s_%s_%d", p.Series, p.Title, p.FirstIssue), p)
		if err != nil {
			break
		}
	}
	return err

}

func (s *Storage) ReadStories() []api.Story {
	records, err := s.db.ReadAll("stories_list")
	if err != nil {
		println(err.Error())
		return make([]api.Story, 0)
	}
	stories := make([]api.Story, 0, len(records))
	for _, p := range records {
		readStory := api.Story{}
		if err := json.Unmarshal(p, &readStory); err != nil {
			fmt.Println("Error", err)
		}
		stories = append(stories, readStory)
	}
	return stories
}

func (s *Storage) ReadKnownTitles() []string {
	records, err := s.db.ReadAll("known_titles")
	if err != nil {
		println(err.Error())
		return defaultKnownTitles
	}
	stories := make([]string, 0, len(records))
	for _, p := range records {
		readStory := ""
		if err := json.Unmarshal(p, &readStory); err != nil {
			fmt.Println("Error", err)
		}
		stories = append(stories, readStory)
	}
	return stories
}

func (s *Storage) ReadSkipTitles() []string {
	records, err := s.db.ReadAll("skip_titles")
	if err != nil {
		println(err.Error())
		return defaultSkipTitles
	}
	stories := make([]string, 0, len(records))
	for _, p := range records {
		readStory := ""
		if err := json.Unmarshal(p, &readStory); err != nil {
			fmt.Println("Error", err)
		}
		stories = append(stories, readStory)
	}
	return stories
}

func NewStorage(storageRoot string) *Storage {
	db, err := scribble.New(storageRoot, nil)
	if err != nil {
		println("Could not connect to storage")
	}
	return &Storage{storageDir: storageRoot, db: db}
}
