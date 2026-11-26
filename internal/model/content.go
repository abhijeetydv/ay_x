package model

// Link represents a labeled hyperlink.
type Link struct {
	Label string `yaml:"label"`
	Href  string `yaml:"href"`
	Page  string `yaml:"page,omitempty"`
}

// NavConfig holds navbar brand and link items.
type NavConfig struct {
	Logo     string `yaml:"logo"`
	LogoHref string `yaml:"logo_href"`
	Links    []Link `yaml:"links"`
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
	Title    string           `yaml:"title"`
	Lang     string           `yaml:"lang"`
	Nav      NavConfig        `yaml:"nav"`
	Sections []SectionConfig  `yaml:"sections"`
	GitHub   GitHubRepoConfig `yaml:"github"`
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

// SocialConfig represents content/social.yaml.
type SocialConfig struct {
	HeroLinks   []SocialLink `yaml:"hero_links"`
	FooterLinks []SocialLink `yaml:"footer_links"`
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
	Title     string `yaml:"title"`
	Heading   string `yaml:"heading"`
	Intro     string `yaml:"intro"`
	BackLabel string `yaml:"back_label"`
	BackHref  string `yaml:"back_href"`
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
	CTAHref     string `yaml:"cta_href"`
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

// PostItem represents a blog post entry on the homepage.
type PostItem struct {
	Title string `yaml:"title"`
	Href  string `yaml:"href"`
	Date  string `yaml:"date"`
}

// PostsCardConfig configures the Latest Posts card.
type PostsCardConfig struct {
	Title        string     `yaml:"title"`
	ExternalHref string     `yaml:"external_href"`
	Items        []PostItem `yaml:"items"`
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
	Posts           PostsCardConfig       `yaml:"posts"`
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
	Copyright string             `yaml:"copyright"`
	Status    string             `yaml:"status"`
	Commit    FooterCommitConfig `yaml:"commit"`
}

// HomepageData aggregates all validated YAML content for template rendering.
type HomepageData struct {
	Site             SiteConfig
	Hero             HeroConfig
	Social           SocialConfig
	Experience       ExperienceConfig
	Projects         ProjectsConfig
	FeaturedProjects []ProjectItem
	AllProjects      []ProjectItem
	Skills           SkillsConfig
	About            AboutConfig
	Footer           FooterConfig
	OrderedSections  []string
}
