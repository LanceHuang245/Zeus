package models

// FilteredSearchResult contains the public city search response fields
type FilteredSearchResult struct {
	Name    string `json:"name"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Address struct {
		State   string `json:"state"`
		Country string `json:"country"`
	} `json:"address"`
}
