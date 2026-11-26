package validate

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"portfolio/internal/model"
)

var (
	validSectionKeys = map[string]bool{
		"hero":           true,
		"experience_bar": true,
		"career":         true,
		"projects":       true,
		"info_grid":      true,
		"bottom_grid":    true,
	}
	validPlatforms = map[string]bool{
		"github":    true,
		"linkedin":  true,
		"x":         true,
		"instagram": true,
	}
	hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{3,8}$`)
	percentRe  = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?%$`)
)

// Homepage validates all loaded YAML content and prepares derived slices
// (OrderedSections and FeaturedProjects).
func Homepage(data *model.HomepageData) error {
	var errs []string

	validateSite(&data.Site, &errs)
	validateHero(&data.Hero, &errs)
	validateSocial(&data.Social, &errs)
	validateExperience(&data.Experience, &errs)
	validateProjects(&data.Projects, &errs)
	validateSkills(&data.Skills, &errs)
	validateAbout(&data.About, &errs)
	validateAboutPage(&data.AboutPage, &errs)
	validatePosts(&data.Posts, &errs)
	validateFooter(&data.Footer, &errs)

	if len(errs) > 0 {
		return errors.New("content validation failed:\n  - " + strings.Join(errs, "\n  - "))
	}

	// Build ordered, enabled section keys.
	if len(data.Site.Sections) == 0 {
		data.OrderedSections = []string{"hero", "experience_bar", "career", "projects", "info_grid", "bottom_grid"}
	} else {
		secs := make([]model.SectionConfig, 0, len(data.Site.Sections))
		for _, s := range data.Site.Sections {
			if s.Enabled {
				secs = append(secs, s)
			}
		}
		sort.SliceStable(secs, func(i, j int) bool {
			return secs[i].Order < secs[j].Order
		})
		for _, s := range secs {
			data.OrderedSections = append(data.OrderedSections, s.Key)
		}
	}

	// Build sorted all projects and featured projects lists.
	allProjects := make([]model.ProjectItem, len(data.Projects.Items))
	copy(allProjects, data.Projects.Items)
	sort.SliceStable(allProjects, func(i, j int) bool {
		return allProjects[i].Order < allProjects[j].Order
	})
	data.AllProjects = allProjects

	featured := make([]model.ProjectItem, 0, len(allProjects))
	for _, p := range allProjects {
		if p.Featured {
			featured = append(featured, p)
		}
	}
	data.FeaturedProjects = featured

	return nil
}

func validateSite(s *model.SiteConfig, errs *[]string) {
	if strings.TrimSpace(s.Title) == "" {
		*errs = append(*errs, "site.yaml: title is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		*errs = append(*errs, "site.yaml: name is required")
	}
	if strings.TrimSpace(s.Lang) == "" {
		s.Lang = "en"
	}
	if strings.TrimSpace(s.Locale) == "" {
		*errs = append(*errs, "site.yaml: locale is required")
	}
	if !isSafeBaseURL(s.URL) {
		*errs = append(*errs, fmt.Sprintf("site.yaml: url %q must be a valid https URL ending with '/'", s.URL))
	}
	validateDescription("site.yaml: description", s.Description, errs)
	if err := validateAssetImage(s.OGImage); err != nil {
		*errs = append(*errs, fmt.Sprintf("site.yaml: og_image %v", err))
	}
	if strings.TrimSpace(s.OGImageAlt) == "" {
		*errs = append(*errs, "site.yaml: og_image_alt is required")
	}
	if err := validateAssetImage(s.Favicon); err != nil {
		*errs = append(*errs, fmt.Sprintf("site.yaml: favicon %v", err))
	}
	if s.ThemeColor != "" && !hexColorRe.MatchString(s.ThemeColor) {
		*errs = append(*errs, fmt.Sprintf("site.yaml: theme_color %q must be a hex color", s.ThemeColor))
	}
	if strings.TrimSpace(s.Nav.Logo) == "" {
		*errs = append(*errs, "site.yaml: nav.logo is required")
	}
	if s.Nav.LogoHref != "" && !isSafeHref(s.Nav.LogoHref) {
		*errs = append(*errs, fmt.Sprintf("site.yaml: invalid nav.logo_href %q", s.Nav.LogoHref))
	}
	for _, requiredKey := range []string{"home", "about", "projects", "posts"} {
		if _, ok := s.Nav.PagePaths[requiredKey]; !ok {
			*errs = append(*errs, fmt.Sprintf("site.yaml: nav.page_paths.%s is required", requiredKey))
		}
	}
	for key, seg := range s.Nav.PagePaths {
		if key != "home" && seg == "" {
			*errs = append(*errs, fmt.Sprintf("site.yaml: nav.page_paths.%s must not be empty", key))
			continue
		}
		if strings.ContainsAny(seg, "/ \t\r\n") {
			*errs = append(*errs, fmt.Sprintf("site.yaml: nav.page_paths.%s %q must not contain '/' or whitespace", key, seg))
		}
	}
	for i, l := range s.Nav.Links {
		if strings.TrimSpace(l.Label) == "" {
			*errs = append(*errs, fmt.Sprintf("site.yaml: nav.links[%d].label is required", i))
		}
		if !isSafeHref(l.Href) {
			*errs = append(*errs, fmt.Sprintf("site.yaml: nav.links[%d].href %q is invalid", i, l.Href))
		}
	}
	seenSections := map[string]bool{}
	for i, sec := range s.Sections {
		if !validSectionKeys[sec.Key] {
			*errs = append(*errs, fmt.Sprintf("site.yaml: sections[%d].key %q is unknown", i, sec.Key))
		}
		if seenSections[sec.Key] {
			*errs = append(*errs, fmt.Sprintf("site.yaml: duplicate section key %q", sec.Key))
		}
		seenSections[sec.Key] = true
	}
}

func validateHero(h *model.HeroConfig, errs *[]string) {
	if strings.TrimSpace(h.Greeting) == "" {
		*errs = append(*errs, "hero.yaml: greeting is required")
	}
	if strings.TrimSpace(h.Name) == "" {
		*errs = append(*errs, "hero.yaml: name is required")
	}
	if len(h.BioParts) == 0 {
		*errs = append(*errs, "hero.yaml: bio_parts must not be empty")
	}
	for i, part := range h.BioParts {
		if part.Href != "" && !isSafeHref(part.Href) {
			*errs = append(*errs, fmt.Sprintf("hero.yaml: bio_parts[%d].href %q is invalid", i, part.Href))
		}
	}
	if h.MoreLink.Label != "" && !isSafeHref(h.MoreLink.Href) {
		*errs = append(*errs, fmt.Sprintf("hero.yaml: more_link.href %q is invalid", h.MoreLink.Href))
	}
}

func validateSocial(s *model.SocialConfig, errs *[]string) {
	check := func(group string, links []model.SocialLink) {
		for i, l := range links {
			if !validPlatforms[strings.ToLower(l.Platform)] {
				*errs = append(*errs, fmt.Sprintf("social.yaml: %s[%d].platform %q is unsupported", group, i, l.Platform))
			}
			if strings.TrimSpace(l.Label) == "" {
				*errs = append(*errs, fmt.Sprintf("social.yaml: %s[%d].label is required", group, i))
			}
			if !isSafeExternalURL(l.URL) {
				*errs = append(*errs, fmt.Sprintf("social.yaml: %s[%d].url %q must be a valid http(s) URL", group, i, l.URL))
			}
		}
	}
	check("hero_links", s.HeroLinks)
	check("footer_links", s.FooterLinks)
}

func validateExperience(e *model.ExperienceConfig, errs *[]string) {
	for i, b := range e.Bar {
		if strings.TrimSpace(b.Abbr) == "" || strings.TrimSpace(b.Name) == "" {
			*errs = append(*errs, fmt.Sprintf("experience.yaml: bar[%d] requires abbr and name", i))
		}
	}
	if strings.TrimSpace(e.Tabs.AriaLabel) == "" {
		*errs = append(*errs, "experience.yaml: tabs.aria_label is required")
	}
	if strings.TrimSpace(e.Tabs.WorkLabel) == "" {
		*errs = append(*errs, "experience.yaml: tabs.work_label is required")
	}
	if strings.TrimSpace(e.Tabs.EducationLabel) == "" {
		*errs = append(*errs, "experience.yaml: tabs.education_label is required")
	}
	checkEntries := func(group string, entries []model.CareerEntry) {
		for i, item := range entries {
			if strings.TrimSpace(item.Role) == "" || strings.TrimSpace(item.Org) == "" {
				*errs = append(*errs, fmt.Sprintf("experience.yaml: tabs.%s[%d] requires role and org", group, i))
			}
		}
	}
	checkEntries("work", e.Tabs.Work)
	checkEntries("education", e.Tabs.Education)
}

func validateProjects(p *model.ProjectsConfig, errs *[]string) {
	if strings.TrimSpace(p.Title) == "" {
		*errs = append(*errs, "projects.yaml: title is required")
	}
	if strings.TrimSpace(p.ViewAll.Label) == "" {
		*errs = append(*errs, "projects.yaml: view_all.label is required")
	}
	if !isSafeHref(p.ViewAll.Href) {
		*errs = append(*errs, fmt.Sprintf("projects.yaml: view_all.href %q is invalid", p.ViewAll.Href))
	}
	if strings.TrimSpace(p.Page.Title) == "" {
		*errs = append(*errs, "projects.yaml: page.title is required")
	}
	validateDescription("projects.yaml: page.description", p.Page.Description, errs)
	if strings.TrimSpace(p.Page.Heading) == "" {
		*errs = append(*errs, "projects.yaml: page.heading is required")
	}
	if strings.TrimSpace(p.Page.Intro) == "" {
		*errs = append(*errs, "projects.yaml: page.intro is required")
	}
	if strings.TrimSpace(p.Page.BackLabel) == "" {
		*errs = append(*errs, "projects.yaml: page.back_label is required")
	}
	if !isSafeHref(p.Page.BackHref) {
		*errs = append(*errs, fmt.Sprintf("projects.yaml: page.back_href %q is invalid", p.Page.BackHref))
	}
	seenIDs := map[string]bool{}
	for i, item := range p.Items {
		if strings.TrimSpace(item.ID) == "" {
			*errs = append(*errs, fmt.Sprintf("projects.yaml: items[%d].id is required", i))
		} else if seenIDs[item.ID] {
			*errs = append(*errs, fmt.Sprintf("projects.yaml: duplicate project id %q", item.ID))
		}
		seenIDs[item.ID] = true
		if strings.TrimSpace(item.Title) == "" {
			*errs = append(*errs, fmt.Sprintf("projects.yaml: items[%d].title is required", i))
		}
		if strings.TrimSpace(item.Owner) == "" || strings.TrimSpace(item.Repo) == "" {
			*errs = append(*errs, fmt.Sprintf("projects.yaml: items[%d] requires owner and repo", i))
		}
		if item.Stars < 0 {
			*errs = append(*errs, fmt.Sprintf("projects.yaml: items[%d].stars cannot be negative", i))
		}
	}
}

func validateSkills(s *model.SkillsConfig, errs *[]string) {
	if strings.TrimSpace(s.Title) == "" {
		*errs = append(*errs, "skills.yaml: title is required")
	}
	if len(s.Items) == 0 {
		*errs = append(*errs, "skills.yaml: items must not be empty")
	}
}

func validateAbout(a *model.AboutConfig, errs *[]string) {
	if strings.TrimSpace(a.Connect.Title) == "" {
		*errs = append(*errs, "about.yaml: connect.title is required")
	}
	if a.Connect.CTAHref != "" && !isSafeHref(a.Connect.CTAHref) {
		*errs = append(*errs, fmt.Sprintf("about.yaml: connect.cta_href %q is invalid", a.Connect.CTAHref))
	}
	if strings.TrimSpace(a.Location.Title) == "" || strings.TrimSpace(a.Location.Name) == "" {
		*errs = append(*errs, "about.yaml: location.title and location.name are required")
	}
	if err := validateAssetImage(a.Location.MapImage); err != nil {
		*errs = append(*errs, fmt.Sprintf("about.yaml: location.map_image %v", err))
	}
	if strings.TrimSpace(a.Location.MapAlt) == "" {
		*errs = append(*errs, "about.yaml: location.map_alt is required")
	}
	if a.GitHubCard.ViewHref != "" && !isSafeHref(a.GitHubCard.ViewHref) {
		*errs = append(*errs, fmt.Sprintf("about.yaml: github_card.view_href %q is invalid", a.GitHubCard.ViewHref))
	}
	for i, lang := range a.GitHubCard.FallbackLanguages {
		if !percentRe.MatchString(lang.Width) {
			*errs = append(*errs, fmt.Sprintf("about.yaml: github_card.fallback_languages[%d].width %q must be a percentage", i, lang.Width))
		}
		if !hexColorRe.MatchString(lang.Color) {
			*errs = append(*errs, fmt.Sprintf("about.yaml: github_card.fallback_languages[%d].color %q must be a hex color", i, lang.Color))
		}
	}
}

func validateAboutPage(a *model.AboutPageConfig, errs *[]string) {
	if strings.TrimSpace(a.Title) == "" {
		*errs = append(*errs, "about_page.yaml: title is required")
	}
	validateDescription("about_page.yaml: description", a.Description, errs)
	if strings.TrimSpace(a.Heading) == "" {
		*errs = append(*errs, "about_page.yaml: heading is required")
	}
	if err := validateAssetImage(a.Photo); err != nil {
		*errs = append(*errs, fmt.Sprintf("about_page.yaml: photo %v", err))
	}
	if strings.TrimSpace(a.PhotoAlt) == "" {
		*errs = append(*errs, "about_page.yaml: photo_alt is required")
	}
	if len(a.Paragraphs) == 0 {
		*errs = append(*errs, "about_page.yaml: paragraphs must not be empty")
	}
	for i, p := range a.Paragraphs {
		if strings.TrimSpace(p) == "" {
			*errs = append(*errs, fmt.Sprintf("about_page.yaml: paragraphs[%d] must not be empty", i))
		}
	}
	for i, f := range a.Facts {
		if strings.TrimSpace(f.Label) == "" || strings.TrimSpace(f.Value) == "" {
			*errs = append(*errs, fmt.Sprintf("about_page.yaml: facts[%d] requires label and value", i))
		}
	}
	if strings.TrimSpace(a.BackLabel) == "" {
		*errs = append(*errs, "about_page.yaml: back_label is required")
	}
	if !isSafeHref(a.BackHref) {
		*errs = append(*errs, fmt.Sprintf("about_page.yaml: back_href %q is invalid", a.BackHref))
	}
	if strings.TrimSpace(a.ProjectsLabel) == "" {
		*errs = append(*errs, "about_page.yaml: projects_label is required")
	}
	if !isSafeHref(a.ProjectsHref) {
		*errs = append(*errs, fmt.Sprintf("about_page.yaml: projects_href %q is invalid", a.ProjectsHref))
	}
}

func validatePosts(p *model.PostsConfig, errs *[]string) {
	if strings.TrimSpace(p.Page.Title) == "" {
		*errs = append(*errs, "posts.yaml: page.title is required")
	}
	validateDescription("posts.yaml: page.description", p.Page.Description, errs)
	if strings.TrimSpace(p.Page.Heading) == "" {
		*errs = append(*errs, "posts.yaml: page.heading is required")
	}
	if strings.TrimSpace(p.Page.Intro) == "" {
		*errs = append(*errs, "posts.yaml: page.intro is required")
	}
	if strings.TrimSpace(p.Page.BackLabel) == "" {
		*errs = append(*errs, "posts.yaml: page.back_label is required")
	}
	if !isSafeHref(p.Page.BackHref) {
		*errs = append(*errs, fmt.Sprintf("posts.yaml: page.back_href %q is invalid", p.Page.BackHref))
	}
	if strings.TrimSpace(p.Page.EmptyMessage) == "" {
		*errs = append(*errs, "posts.yaml: page.empty_message is required")
	}
	if strings.TrimSpace(p.Page.DateFormat) == "" {
		*errs = append(*errs, "posts.yaml: page.date_format is required")
	}
	if strings.TrimSpace(p.HomeTile.Title) == "" {
		*errs = append(*errs, "posts.yaml: home_tile.title is required")
	}
	if strings.TrimSpace(p.HomeTile.AllPostsLabel) == "" {
		*errs = append(*errs, "posts.yaml: home_tile.all_posts_label is required")
	}
	if !isSafeHref(p.HomeTile.AllPostsHref) {
		*errs = append(*errs, fmt.Sprintf("posts.yaml: home_tile.all_posts_href %q is invalid", p.HomeTile.AllPostsHref))
	}
	seenIDs := map[string]bool{}
	for i, item := range p.Items {
		if strings.TrimSpace(item.ID) == "" {
			*errs = append(*errs, fmt.Sprintf("posts.yaml: items[%d].id is required", i))
		} else if seenIDs[item.ID] {
			*errs = append(*errs, fmt.Sprintf("posts.yaml: duplicate post id %q", item.ID))
		}
		seenIDs[item.ID] = true
		if strings.TrimSpace(item.Title) == "" {
			*errs = append(*errs, fmt.Sprintf("posts.yaml: items[%d].title is required", i))
		}
		if _, err := time.Parse("2006-01-02", item.Date); err != nil {
			*errs = append(*errs, fmt.Sprintf("posts.yaml: items[%d].date %q must be YYYY-MM-DD", i, item.Date))
		}
		if !isSafeHref(item.URL) {
			*errs = append(*errs, fmt.Sprintf("posts.yaml: items[%d].url %q is invalid", i, item.URL))
		}
	}
}

func validateFooter(f *model.FooterConfig, errs *[]string) {
	if strings.TrimSpace(f.Copyright) == "" {
		*errs = append(*errs, "footer.yaml: copyright is required")
	}
	if strings.TrimSpace(f.TimerLabel) == "" {
		*errs = append(*errs, "footer.yaml: timer_label is required")
	}
	if f.Commit.URL != "" && !isSafeExternalURL(f.Commit.URL) {
		*errs = append(*errs, fmt.Sprintf("footer.yaml: commit.url %q must be a valid http(s) URL", f.Commit.URL))
	}
}

func isSafeHref(raw string) bool {
	if raw == "" {
		return false
	}
	if raw == "#" || strings.HasPrefix(raw, "#") {
		return true
	}
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.Contains(raw, "..") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return u.Host == ""
	}
	return scheme == "http" || scheme == "https" || scheme == "mailto" || scheme == "tel"
}

func isSafeExternalURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https"
}

func isSafeBaseURL(raw string) bool {
	if !strings.HasSuffix(raw, "/") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	return strings.ToLower(u.Scheme) == "https"
}

func validateDescription(field, desc string, errs *[]string) {
	trimmed := strings.TrimSpace(desc)
	if trimmed == "" {
		*errs = append(*errs, field+" is required")
		return
	}
	if utf8.RuneCountInString(trimmed) > 200 {
		*errs = append(*errs, field+" must be at most 200 characters")
	}
}

func validateAssetImage(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return errors.New("is required")
	}
	if strings.Contains(trimmed, "://") || strings.HasPrefix(trimmed, "/") || filepath.IsAbs(trimmed) {
		return fmt.Errorf("%q must be a relative path under assets/", raw)
	}
	if strings.Contains(trimmed, "..") {
		return fmt.Errorf("%q must not contain '..'", raw)
	}
	cleaned := filepath.ToSlash(filepath.Clean(trimmed))
	if !strings.HasPrefix(cleaned, "assets/") || cleaned == "assets/" {
		return fmt.Errorf("%q must point to a file under assets/", raw)
	}
	resolved, ok := resolveAssetPath(cleaned)
	if !ok {
		return fmt.Errorf("%q does not exist under assets/", raw)
	}
	info, err := os.Stat(resolved)
	if err != nil || info.IsDir() {
		return fmt.Errorf("%q is not a valid file", raw)
	}
	return nil
}

func resolveAssetPath(rel string) (string, bool) {
	if _, err := os.Stat(rel); err == nil {
		return rel, true
	}
	for _, prefix := range []string{"..", filepath.Join("..", "..")} {
		if _, err := os.Stat(filepath.Join(prefix, "go.mod")); err == nil {
			candidate := filepath.Join(prefix, rel)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, true
			}
		}
	}
	return "", false
}
