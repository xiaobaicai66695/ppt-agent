package unsplash

// SearchOptions contains the public /search/photos parameters.
type SearchOptions struct {
	Query         string
	Orientation   string
	ContentFilter string
	Color         string
	OrderBy       string
	Page          int
	PerPage       int
}

// SearchResponse is the abbreviated photo search response returned by Unsplash.
type SearchResponse struct {
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
	Results    []Photo `json:"results"`
}

// Photo is the subset of an Unsplash photo object needed by the asset workflow.
type Photo struct {
	ID             string     `json:"id"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	Color          string     `json:"color"`
	Description    string     `json:"description"`
	AltDescription string     `json:"alt_description"`
	BlurHash       string     `json:"blur_hash"`
	URLs           PhotoURLs  `json:"urls"`
	Links          PhotoLinks `json:"links"`
	User           User       `json:"user"`
}

type PhotoURLs struct {
	Raw     string `json:"raw"`
	Full    string `json:"full"`
	Regular string `json:"regular"`
	Small   string `json:"small"`
	Thumb   string `json:"thumb"`
}

type PhotoLinks struct {
	HTML             string `json:"html"`
	Download         string `json:"download"`
	DownloadLocation string `json:"download_location"`
}

type User struct {
	Name     string    `json:"name"`
	Username string    `json:"username"`
	Links    UserLinks `json:"links"`
}

type UserLinks struct {
	HTML string `json:"html"`
}
