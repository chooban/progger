package models

import (
	"github.com/rushysloth/go-tsid"
)

type Library struct {
	ID                                int64    `db:"id"`
	Name                              string   `db:"name"`
	Root                              string   `db:"root"`
	OneshotsDirectory                 string   `db:"oneshots_directory"`
	AnalyzeDimensions                 bool     `db:"analyze_dimensions"`
	ConvertToCbz                      bool     `db:"convert_to_cbz"`
	EmptyTrashAfterScan               bool     `db:"empty_trash_after_scan"`
	HashFiles                         bool     `db:"hash_files"`
	HashKoreader                      bool     `db:"hash_koreader"`
	HashPages                         bool     `db:"hash_pages"`
	ImportBarcodeIsbn                 bool     `db:"import_barcode_isbn"`
	ImportComicInfoBook               bool     `db:"import_comic_info_book"`
	ImportComicInfoCollection         bool     `db:"import_comic_info_collection"`
	ImportComicInfoReadList           bool     `db:"import_comic_info_read_list"`
	ImportComicInfoSeries             bool     `db:"import_comic_info_series"`
	ImportComicInfoSeriesAppendVolume bool     `db:"import_comic_info_series_append_volume"`
	ImportEpubBook                    bool     `db:"import_epub_book"`
	ImportEpubSeries                  bool     `db:"import_epub_series"`
	ImportLocalArtwork                bool     `db:"import_local_artwork"`
	ImportMylarSeries                 bool     `db:"import_mylar_series"`
	RepairExtensions                  bool     `db:"repair_extensions"`
	ScanCbx                           bool     `db:"scan_cbx"`
	ScanEpub                          bool     `db:"scan_epub"`
	ScanPdf                           bool     `db:"scan_pdf"`
	ScanOnStartup                     bool     `db:"scan_on_startup"`
	ScanForceModifiedTime             bool     `db:"scan_force_modified_time"`
	ScanInterval                      string   `db:"scan_interval"`
	ScanDirectoryExclusions           []string `db:"scan_directory_exclusions"`
	SeriesCover                       string   `db:"series_cover"`
	Unavailable                       bool     `db:"unavailable"`
	CreatedAt                         string   `db:"created_at"`
	UpdatedAt                         string   `db:"updated_at"`
}

func (l Library) URL() string {
	t := tsid.FromNumber(l.ID)
	return "/api/v1/libraries/" + t.ToString()
}
