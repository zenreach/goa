package model

// List of blogs for user
type BlogList struct {
	Blogs     []*Blog         `attribute:"blogs"`        // The list of Blogs this user has Authorship or Admin rights for.
	UserInfos []*BlogUserInfo `attribute:"blogUserInfo"` // Admin level list of blog per-user information
}

// Blog
// Note how structures can recurse (Posts, Pages and Locale)
type Blog struct {
	Id          string    `attribute:"id"`
	Name        string    `attribute:"name"`
	Description string    `attribute:"description"`
	Published   time.Time `attribute:"published"`
	Updated     time.Time `attribute:"published"`
	Url         string    `attribute:"url"`
	SelfLink    string    `attribute:"selfLink"`
	Posts       *struct {
		TotalItems int    `attribute:"totalItems"`
		SelfLink   string `attribute:"selfLink"`
	} `attribute:"posts"`
	Pages *struct {
		TotalItems int    `attribute:"totalItems"`
		SelfLink   string `attribute:"selfLink"`
	} `attribute:"pages"`
	Locale *struct {
		Language string `attribute:"language"`
		Country  string `attribute:"country"`
		Variant  string `attribute:"variant"`
	} `attribute:"locale"`
}
