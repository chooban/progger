package internal

type PdfBookmark struct {
	Title    string
	PageFrom int
	PageThru int
}

type EpisodeDetails struct {
	Bookmark PdfBookmark
	Credits  string
}
