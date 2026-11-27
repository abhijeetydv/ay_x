package model

import "html/template"

// AllowedTitleColors defines the single source of truth for post title_colors entries:
// red, orange, yellow, green, cyan, blue, purple, pink, white, gray.
// To add a new colour:
//  1. Add its name to AllowedTitleColors here.
//  2. Add --color-<name> and .title-<name> in styles.css (a test enforces this).
//  3. Update the example comment in content/posts/*.md if desired.
var AllowedTitleColors = []string{
	"red",
	"orange",
	"yellow",
	"green",
	"cyan",
	"blue",
	"purple",
	"pink",
	"white",
	"gray",
}

// Link represents a labeled hyperlink.
type Link struct {
	Label  string `yaml:"label"`
	Href   string `yaml:"href"`
	Page   string `yaml:"page,omitempty"`
	NewTab bool   `yaml:"new_tab,omitempty"`
}

// NavConfig holds navbar brand and link items.
type NavConfig struct {
	Logo           string            `yaml:"logo"`
	LogoHref       string            `yaml:"logo_href"`
	NewTabLabel    string            `yaml:"new_tab_label,omitempty"`
	MenuOpenLabel  string            `yaml:"menu_open_label"`
	MenuCloseLabel string            `yaml:"menu_close_label"`
	PanelAriaLabel string            `yaml:"panel_aria_label"`
	PagePaths      map[string]string `yaml:"page_paths"`
	Links          []Link            `yaml:"links"`
}

// SectionConfig controls homepage section visibility and ordering.
type SectionConfig struct {
	Key     string `yaml:"key"`
	Enabled bool   `yaml:"enabled"`
	Order   int    `yaml:"order"`
}

// GitHubRepoConfig identifies the GitHub user and repository for live stats.
type GitHubRepoConfig struct {
	Username string `yaml:"username"`
	Repo     string `yaml:"repo"`
}

// SiteConfig represents content/site.yaml.
type SiteConfig struct {
	Title       string           `yaml:"title"`
	Name        string           `yaml:"name"`
	Lang        string           `yaml:"lang"`
	Locale      string           `yaml:"locale"`
	URL         string           `yaml:"url"`
	Description string           `yaml:"description"`
	OGImage     string           `yaml:"og_image"`
	OGImageAlt  string           `yaml:"og_image_alt"`
	Favicon     string           `yaml:"favicon"`
	ThemeColor  string           `yaml:"theme_color,omitempty"`
	Nav         NavConfig        `yaml:"nav"`
	Sections    []SectionConfig  `yaml:"sections"`
	GitHub      GitHubRepoConfig `yaml:"github"`
}

// BioPart represents either plain text or an inline-link block in the hero bio.
type BioPart struct {
	Text string `yaml:"text"`
	Href string `yaml:"href,omitempty"`
}

// HeroConfig represents content/hero.yaml.
type HeroConfig struct {
	Greeting string    `yaml:"greeting"`
	Name     string    `yaml:"name"`
	BioParts []BioPart `yaml:"bio_parts"`
	MoreLink Link      `yaml:"more_link"`
}

// SocialLink represents a social profile link.
type SocialLink struct {
	Platform string `yaml:"platform"`
	Label    string `yaml:"label"`
	URL      string `yaml:"url"`
}

// DockEmailConfig configures the email entry in the social dock.
type DockEmailConfig struct {
	Label   string `yaml:"label"`
	Address string `yaml:"address"`
}

// DockConfig configures the footer / floating social dock.
type DockConfig struct {
	AriaLabel   string          `yaml:"aria_label"`
	BubbleTitle string          `yaml:"bubble_title"`
	Email       DockEmailConfig `yaml:"email"`
	Items       []string        `yaml:"items"`
}

// DockItem is a resolved entry ready for rendering in the social dock.
type DockItem struct {
	Platform   string
	Label      string
	Href       string
	IsExternal bool
}

// SocialConfig represents content/social.yaml.
type SocialConfig struct {
	HeroLinks   []SocialLink `yaml:"hero_links"`
	FooterLinks []SocialLink `yaml:"footer_links"`
	Dock        DockConfig   `yaml:"dock"`
}

// ExperienceBarItem represents an item in the compact experience strip.
type ExperienceBarItem struct {
	Abbr   string `yaml:"abbr"`
	Name   string `yaml:"name"`
	Tag    string `yaml:"tag,omitempty"`
	Active bool   `yaml:"active,omitempty"`
}

// CareerEntry represents a single work experience or education entry.
type CareerEntry struct {
	Role        string `yaml:"role"`
	Org         string `yaml:"org"`
	Period      string `yaml:"period"`
	Description string `yaml:"description"`
}

// CareerTabsConfig holds tabbed work experience and education data.
type CareerTabsConfig struct {
	AriaLabel      string        `yaml:"aria_label"`
	WorkLabel      string        `yaml:"work_label"`
	EducationLabel string        `yaml:"education_label"`
	Work           []CareerEntry `yaml:"work"`
	Education      []CareerEntry `yaml:"education"`
}

// ExperienceConfig represents content/experience.yaml.
type ExperienceConfig struct {
	Bar  []ExperienceBarItem `yaml:"bar"`
	Tabs CareerTabsConfig    `yaml:"tabs"`
}

// ProjectItem represents a project card.
type ProjectItem struct {
	ID           string   `yaml:"id"`
	Owner        string   `yaml:"owner"`
	Repo         string   `yaml:"repo"`
	Stars        int      `yaml:"stars"`
	TerminalDesc string   `yaml:"terminal_desc"`
	Title        string   `yaml:"title"`
	Summary      string   `yaml:"summary"`
	Tags         []string `yaml:"tags"`
	Featured     bool     `yaml:"featured"`
	Order        int      `yaml:"order"`
}

// ProjectsPageConfig holds metadata and copy for projects.html.
type ProjectsPageConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Heading     string `yaml:"heading"`
	Intro       string `yaml:"intro"`
	BackLabel   string `yaml:"back_label"`
	BackHref    string `yaml:"back_href"`
}

// ProjectsConfig represents content/projects.yaml.
type ProjectsConfig struct {
	Title   string             `yaml:"title"`
	Icon    string             `yaml:"icon"`
	ViewAll Link               `yaml:"view_all"`
	Page    ProjectsPageConfig `yaml:"page"`
	Items   []ProjectItem      `yaml:"items"`
}

// SkillsConfig represents content/skills.yaml.
type SkillsConfig struct {
	Title string   `yaml:"title"`
	Items []string `yaml:"items"`
}

// ConnectCardConfig configures the Let's Connect tile.
type ConnectCardConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	CTALabel    string `yaml:"cta_label"`
}

// LocationCardConfig configures the Currently Based In tile.
type LocationCardConfig struct {
	Title    string `yaml:"title"`
	Name     string `yaml:"name"`
	MapLabel string `yaml:"map_label"`
	MapImage string `yaml:"map_image"`
	MapAlt   string `yaml:"map_alt"`
	Timezone string `yaml:"timezone"`
}

// PlaceholderTileConfig configures the reserved 4th tile.
type PlaceholderTileConfig struct {
	Enabled bool `yaml:"enabled"`
}

// PostsPageConfig configures posts.html metadata and copy.
type PostsPageConfig struct {
	Title        string `yaml:"title"`
	Description  string `yaml:"description"`
	Heading      string `yaml:"heading"`
	Intro        string `yaml:"intro"`
	BackLabel    string `yaml:"back_label"`
	BackHref     string `yaml:"back_href"`
	EmptyMessage string `yaml:"empty_message"`
	DateFormat   string `yaml:"date_format"`
}

// PostsHomeTileConfig configures the home page latest-post tile.
type PostsHomeTileConfig struct {
	Title         string `yaml:"title"`
	AllPostsLabel string `yaml:"all_posts_label"`
	AllPostsHref  string `yaml:"all_posts_href"`
}

// PostPageConfig configures individual post pages (posts/<slug>/index.html).
type PostPageConfig struct {
	BackLabel         string `yaml:"back_label"`
	BackHref          string `yaml:"back_href"`
	UpdatedLabel      string `yaml:"updated_label"`
	CopyLabel         string `yaml:"copy_label"`
	CopiedLabel       string `yaml:"copied_label"`
	ContentsLabel     string `yaml:"contents_label"`
	ContentsAriaLabel string `yaml:"contents_aria_label"`
	TOCMinItems       int    `yaml:"toc_min_items"`
}

// TOCItem represents a heading entry in a post's table of contents.
type TOCItem struct {
	ID       string
	Text     string
	Level    int
	Children []TOCItem
}

// TitleColor represents a coloured substring rule from post front matter.
type TitleColor struct {
	Text  string
	Color string
}

// TitleSegment represents an ordered plain or coloured segment of a post title.
type TitleSegment struct {
	Text  string
	Color string
}

// Post represents a single Markdown blog post loaded from content/posts/*.md.
type Post struct {
	Slug             string
	SourcePath       string
	Title            string
	TitleColors      []TitleColor
	TitleSegments    []TitleSegment
	Date             string
	FormattedDate    string
	Updated          string
	FormattedUpdated string
	Summary          string
	Tags             []string
	RawBody          string
	Body             template.HTML
	TOC              []TOCItem
	ShowTOC          bool
	URL              string
}

// PostsConfig represents content/posts.yaml.
type PostsConfig struct {
	Page     PostsPageConfig     `yaml:"page"`
	HomeTile PostsHomeTileConfig `yaml:"home_tile"`
	PostPage PostPageConfig      `yaml:"post_page"`
}

// FallbackCommit represents a fallback commit row in the GitHub Activity card.
type FallbackCommit struct {
	Repo    string `yaml:"repo"`
	Message string `yaml:"message"`
	SHA     string `yaml:"sha"`
}

// FallbackLanguageBar represents a fallback language bar segment.
type FallbackLanguageBar struct {
	Width string `yaml:"width"`
	Color string `yaml:"color"`
}

// GitHubCardConfig configures the GitHub Activity card.
type GitHubCardConfig struct {
	Title             string                `yaml:"title"`
	InfoLabel         string                `yaml:"info_label"`
	InfoHref          string                `yaml:"info_href"`
	ViewLabel         string                `yaml:"view_label"`
	ViewHref          string                `yaml:"view_href"`
	FallbackCommits   []FallbackCommit      `yaml:"fallback_commits"`
	FallbackLanguages []FallbackLanguageBar `yaml:"fallback_languages"`
}

// AboutConfig represents content/about.yaml.
type AboutConfig struct {
	Connect         ConnectCardConfig     `yaml:"connect"`
	Location        LocationCardConfig    `yaml:"location"`
	PlaceholderTile PlaceholderTileConfig `yaml:"placeholder_tile"`
	GitHubCard      GitHubCardConfig      `yaml:"github_card"`
}

// FooterCommitConfig configures the footer commit badge.
type FooterCommitConfig struct {
	SHA   string `yaml:"sha"`
	URL   string `yaml:"url"`
	Title string `yaml:"title"`
}

// FooterConfig represents content/footer.yaml.
type FooterConfig struct {
	Copyright  string             `yaml:"copyright"`
	Status     string             `yaml:"status"`
	TimerLabel string             `yaml:"timer_label"`
	Commit     FooterCommitConfig `yaml:"commit"`
}

// FactItem represents a label/value fact on the About page.
type FactItem struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// AboutPageConfig represents content/about_page.yaml.
type AboutPageConfig struct {
	Title         string     `yaml:"title"`
	Description   string     `yaml:"description"`
	Heading       string     `yaml:"heading"`
	Photo         string     `yaml:"photo"`
	PhotoAlt      string     `yaml:"photo_alt"`
	Paragraphs    []string   `yaml:"paragraphs"`
	Facts         []FactItem `yaml:"facts,omitempty"`
	BackLabel     string     `yaml:"back_label"`
	BackHref      string     `yaml:"back_href"`
	ProjectsLabel string     `yaml:"projects_label"`
	ProjectsHref  string     `yaml:"projects_href"`
}

// HomepageData aggregates all validated YAML content for template rendering.
type HomepageData struct {
	Site             SiteConfig
	Hero             HeroConfig
	Social           SocialConfig
	DockItems        []DockItem
	Experience       ExperienceConfig
	Projects         ProjectsConfig
	FeaturedProjects []ProjectItem
	AllProjects      []ProjectItem
	Skills           SkillsConfig
	About            AboutConfig
	AboutPage        AboutPageConfig
	Posts            PostsConfig
	SortedPosts      []Post
	LatestPost       *Post
	Footer           FooterConfig
	OrderedSections  []string
}
