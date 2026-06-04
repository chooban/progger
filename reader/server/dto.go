package server

type SeriesDto struct {
	ID                   string                     `json:"id"`
	LibraryID            string                     `json:"libraryId"`
	Name                 string                     `json:"name"`
	BookCount            int                        `json:"booksCount"`
	BooksInProgressCount int                        `json:"booksInProgressCount"`
	BooksReadCount       int                        `json:"booksReadCount"`
	BooksUnreadCount     int                        `json:"booksUnreadCount"`
	BooksMetadata        BookMetadataAggregationDto `json:"booksMetadata"`
	URL                  string                     `json:"url"`
	Created              string                     `json:"created"`
	LastModified         string                     `json:"lastModified"`
	FileLastModified     string                     `json:"fileLastModified"`
	Metadata             SeriesMetadataDto          `json:"metadata"`
	Deleted              bool                       `json:"deleted"`
	Oneshot              bool                       `json:"oneshot"`
}

type SeriesMetadataDto struct {
	Title                string              `json:"title"`
	Summary              string              `json:"summary"`
	AgeRating            int32               `json:"ageRating"`
	AgeRatingLock        bool                `json:"ageRatingLock"`
	AlternateTitles      []AlternateTitleDto `json:"alternateTitles"`
	AlternateTitlesLock  bool                `json:"alternateTitlesLock"`
	Created              string              `json:"created"`
	Genres               []string            `json:"genres"`
	GenresLock           bool                `json:"genresLock"`
	Language             string              `json:"language"`
	LanguageLock         bool                `json:"languageLock"`
	LastModified         string              `json:"lastModified"`
	Links                []WebLinkDto        `json:"links"`
	LinksLock            bool                `json:"linksLock"`
	Publisher            string              `json:"publisher"`
	PublisherLock        bool                `json:"publisherLock"`
	ReadingDirection     string              `json:"readingDirection"`
	ReadingDirectionLock bool                `json:"readingDirectionLock"`
	SharingLabels        []string            `json:"sharingLabels"`
	SharingLabelsLock    bool                `json:"sharingLabelsLock"`
	Status               string              `json:"status"`
	StatusLock           bool                `json:"statusLock"`
	Tags                 []string            `json:"tags"`
	TagsLock             bool                `json:"tagsLock"`
	TitleSort            string              `json:"titleSort"`
	TitleSortLock        bool                `json:"titleSortLock"`
	TotalBookCount       int32               `json:"totalBookCount"`
	TotalBookCountLock   bool                `json:"totalBookCountLock"`
}

type AlternateTitleDto struct {
	Title string `json:"title"`
}

type AuthorDto struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type BookMetadataAggregationDto struct {
	Authors       []AuthorDto `json:"authors"`
	Created       string      `json:"created"`
	LastModified  string      `json:"lastModified"`
	Summary       string      `json:"summary"`
	SummaryNumber string      `json:"summaryNumber"`
	Tags          []string    `json:"tags"`
	ReleaseDate   *string     `json:"releaseDate"`
}

type WebLinkDto struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type BookDto struct {
	ID               string           `json:"id"`
	SeriesID         string           `json:"seriesId"`
	SeriesTitle      string           `json:"seriesTitle"`
	Name             string           `json:"name"`
	Number           int32            `json:"number"`
	SizeBytes        int64            `json:"sizeBytes"`
	Size             string           `json:"size"`
	URL              string           `json:"url"`
	Created          string           `json:"created"`
	LastModified     string           `json:"lastModified"`
	LibraryID        string           `json:"libraryId"`
	Deleted          bool             `json:"deleted"`
	FileHash         string           `json:"fileHash"`
	FileLastModified string           `json:"fileLastModified"`
	Oneshot          bool             `json:"oneshot"`
	Media            MediaDto         `json:"media"`
	Metadata         BookMetadataDto  `json:"metadata"`
	ReadProgress     *ReadProgressDto `json:"readProgress"`
}

type MediaDto struct {
	Comment              string `json:"comment"`
	EpubDivinaCompatible bool   `json:"epubDivinaCompatible"`
	EpubIsKepub          bool   `json:"epubIsKepub"`
	MediaProfile         string `json:"mediaProfile"`
	MediaType            string `json:"mediaType"`
	PagesCount           int32  `json:"pagesCount"`
	Status               string `json:"status"`
}

type PageDto struct {
	Number    int    `json:"number"`
	MediaType string `json:"mediaType"`
	FileName  string `json:"fileName"`
	Height    int    `json:"height"`
	Width     int    `json:"width"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"sizeBytes"`
}

type ThumbnailDto struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Selected bool   `json:"selected"`
	URL      string `json:"url"`
}

type LibraryDto struct {
	AnalyzeDimensions                 bool     `json:"analyzeDimensions"`
	ConvertToCbz                      bool     `json:"convertToCbz"`
	EmptyTrashAfterScan               bool     `json:"emptyTrashAfterScan"`
	HashFiles                         bool     `json:"hashFiles"`
	HashKoreader                      bool     `json:"hashKoreader"`
	HashPages                         bool     `json:"hashPages"`
	ID                                string   `json:"id"`
	ImportBarcodeIsbn                 bool     `json:"importBarcodeIsbn"`
	ImportComicInfoBook               bool     `json:"importComicInfoBook"`
	ImportComicInfoCollection         bool     `json:"importComicInfoCollection"`
	ImportComicInfoReadList           bool     `json:"importComicInfoReadList"`
	ImportComicInfoSeries             bool     `json:"importComicInfoSeries"`
	ImportComicInfoSeriesAppendVolume bool     `json:"importComicInfoSeriesAppendVolume"`
	ImportEpubBook                    bool     `json:"importEpubBook"`
	ImportEpubSeries                  bool     `json:"importEpubSeries"`
	ImportLocalArtwork                bool     `json:"importLocalArtwork"`
	ImportMylarSeries                 bool     `json:"importMylarSeries"`
	Name                              string   `json:"name"`
	OneshotsDirectory                 string   `json:"oneshotsDirectory"`
	RepairExtensions                  bool     `json:"repairExtensions"`
	Root                              string   `json:"root"`
	ScanCbx                           bool     `json:"scanCbx"`
	ScanDirectoryExclusions           []string `json:"scanDirectoryExclusions"`
	ScanEpub                          bool     `json:"scanEpub"`
	ScanForceModifiedTime             bool     `json:"scanForceModifiedTime"`
	ScanInterval                      string   `json:"scanInterval"`
	ScanOnStartup                     bool     `json:"scanOnStartup"`
	ScanPdf                           bool     `json:"scanPdf"`
	SeriesCover                       string   `json:"seriesCover"`
	Unavailable                       bool     `json:"unavailable"`
}

type HealthDto struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type UserDto struct {
	Email              string             `json:"email"`
	ID                 string             `json:"id"`
	LabelsAllow        []string           `json:"labelsAllow"`
	LabelsExclude      []string           `json:"labelsExclude"`
	Roles              []string           `json:"roles"`
	SharedAllLibraries bool               `json:"sharedAllLibraries"`
	SharedLibrariesIds []string           `json:"sharedLibrariesIds"`
	AgeRestriction     *AgeRestrictionDto `json:"ageRestriction"`
}

type AgeRestrictionDto struct {
	Age         int32  `json:"age"`
	Restriction string `json:"restriction"`
}

type CollectionDto struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Filtered         bool     `json:"filtered"`
	Ordered          bool     `json:"ordered"`
	SeriesIds        []string `json:"seriesIds"`
	CreatedDate      string   `json:"createdDate"`
	LastModifiedDate string   `json:"lastModifiedDate"`
}

type ReadListDto struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Summary          string   `json:"summary"`
	Filtered         bool     `json:"filtered"`
	Ordered          bool     `json:"ordered"`
	BookIds          []string `json:"bookIds"`
	CreatedDate      string   `json:"createdDate"`
	LastModifiedDate string   `json:"lastModifiedDate"`
}

type BookSearchRequest struct {
	Condition      BookSearchCondition `json:"condition"`
	FullTextSearch string              `json:"fullTextSearch,omitempty"`
}

type BookSearchCondition struct {
	AllOf      []BookSearchCondition `json:"allOf"`
	AnyOf      []BookSearchCondition `json:"anyOf"`
	SeriesId   *ConditionValue       `json:"seriesId"`
	LibraryId  *ConditionValue       `json:"libraryId"`
	ReadStatus *ConditionValue       `json:"readStatus"`
}

type ConditionValue struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func (c BookSearchCondition) IsEmpty() bool {
	return c.AllOf == nil && c.AnyOf == nil && c.SeriesId == nil && c.LibraryId == nil && c.ReadStatus == nil
}

type SeriesSearchRequest struct {
	Condition SeriesSearchCondition `json:"condition"`
}

type SeriesSearchCondition struct {
	AllOf      []SeriesSearchCondition `json:"allOf"`
	AnyOf      []SeriesSearchCondition `json:"anyOf"`
	LibraryId  *ConditionValue         `json:"libraryId"`
	CollectionId *ConditionValue       `json:"collectionId"`
}

func (c SeriesSearchCondition) IsEmpty() bool {
	return c.AllOf == nil && c.AnyOf == nil && c.LibraryId == nil && c.CollectionId == nil
}

type BookMetadataDto struct {
	Authors         []AuthorDto  `json:"authors"`
	AuthorsLock     bool         `json:"authorsLock"`
	Created         string       `json:"created"`
	ISBN            string       `json:"isbn"`
	ISBNLock        bool         `json:"isbnLock"`
	LastModified    string       `json:"lastModified"`
	Links           []WebLinkDto `json:"links"`
	LinksLock       bool         `json:"linksLock"`
	Number          string       `json:"number"`
	NumberLock      bool         `json:"numberLock"`
	NumberSort      float64      `json:"numberSort"`
	NumberSortLock  bool         `json:"numberSortLock"`
	ReleaseDate     *string      `json:"releaseDate"`
	ReleaseDateLock bool         `json:"releaseDateLock"`
	Summary         string       `json:"summary"`
	SummaryLock     bool         `json:"summaryLock"`
	Tags            []string     `json:"tags"`
	TagsLock        bool         `json:"tagsLock"`
	Title           string       `json:"title"`
	TitleLock       bool         `json:"titleLock"`
}

type ReadProgressDto struct {
	Completed    bool   `json:"completed"`
	Created      string `json:"created"`
	DeviceID     string `json:"deviceId"`
	DeviceName   string `json:"deviceName"`
	LastModified string `json:"lastModified"`
	Page         int32  `json:"page"`
	ReadDate     string `json:"readDate"`
}
