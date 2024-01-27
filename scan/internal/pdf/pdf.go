package pdf

type Reader interface {
	Bookmarks(filename string) ([]EpisodeDetails, error)
	//Build(episodes []types.Episode)
	Credits(filename string, startPage int, endPage int) (string, error)
}

type Bookmark struct {
	Title    string
	PageFrom int
	PageThru int
}

type EpisodeDetails struct {
	Bookmark Bookmark
	Credits  string
}
