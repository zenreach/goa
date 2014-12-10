package model

// List of blogs for user
type PostList struct {
	Posts         []*Post `attribute:"items"`
	NextPageToken string  `attribute:"nextPageToken"`
}

// Blog post
type Post struct {
	Id        string    `attribute:"id"`
	Blog      *Blog     `attribute:"BlogBlog"`
	Published time.Time `attribute:"published"`
	Updated   time.Time `attribute:"updated"`
	Url       string    `attribute:"url"`
	SelfLink  string    `attribute:"selfLink"`
	Title     string    `attribute:"title"`
	TitleLink string    `attribute:"titleLink"`
	Content   string    `attribute:"content"`
	Author    *Author   `attribute:"author"`
	Replies   *Replies  `attribute:"replies"`
	Labels    []string  `attribute:"labels"`
	Medatata  string    `attribute:"customMedatata"`
	Location  *Location `attribute:"location"`
	Images    []*Image  `attribute:"images"`
	Status    string    `attribute:"status"`
}
