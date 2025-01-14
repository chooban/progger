package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chooban/progger/exporter/api"
	"github.com/sdomino/scribble"
)

type Storage struct {
	storageDir string
	db         *scribble.Driver
}

func (s *Storage) SaveProgs(progs []api.Downloadable) error {
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

func (s *Storage) ReadProgs() []api.Downloadable {
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

func NewStorage(storageRoot string) *Storage {
	db, err := scribble.New(storageRoot, nil)
	if err != nil {
		println("Could not connect to storage")
	}
	return &Storage{storageDir: storageRoot, db: db}
}
