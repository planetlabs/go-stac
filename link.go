package stac

type Link struct {
	Href  string `json:"href"`
	Rel   string `json:"rel"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}
