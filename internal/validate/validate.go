package validate

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

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

	// Build sorted featured projects list.
	featured := make([]model.ProjectItem, 0, len(data.Projects.Items))
	for _, p := range data.Projects.Items {
		if p.Featured {
			featured = append(featured, p)
		}
	}
	sort.SliceStable(featured, func(i, j int) bool {
		return featured[i].Order < featured[j].Order
	})
	data.FeaturedProjects = featured

	return nil
}

func validateSite(s *model.SiteConfig, errs *[]string) {
	if strings.TrimSpace(s.Title) == "" {
		*errs = append(*errs, "site.yaml: title is required")
	}
	if strings.TrimSpace(s.Lang) == "" {
		s.Lang = "en"
	}
	if strings.TrimSpace(s.Nav.Logo) == "" {
		*errs = append(*errs, "site.yaml: nav.logo is required")
	}
	if s.Nav.LogoHref != "" && !isSafeHref(s.Nav.LogoHref) {
		*errs = append(*errs, fmt.Sprintf("site.yaml: invalid nav.logo_href %q", s.Nav.LogoHref))
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
	if p.ViewAll.Href != "" && !isSafeHref(p.ViewAll.Href) {
		*errs = append(*errs, fmt.Sprintf("projects.yaml: view_all.href %q is invalid", p.ViewAll.Href))
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
	if a.Posts.ExternalHref != "" && !isSafeHref(a.Posts.ExternalHref) {
		*errs = append(*errs, fmt.Sprintf("about.yaml: posts.external_href %q is invalid", a.Posts.ExternalHref))
	}
	for i, post := range a.Posts.Items {
		if strings.TrimSpace(post.Title) == "" {
			*errs = append(*errs, fmt.Sprintf("about.yaml: posts.items[%d].title is required", i))
		}
		if !isSafeHref(post.Href) {
			*errs = append(*errs, fmt.Sprintf("about.yaml: posts.items[%d].href %q is invalid", i, post.Href))
		}
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

func validateFooter(f *model.FooterConfig, errs *[]string) {
	if strings.TrimSpace(f.Copyright) == "" {
		*errs = append(*errs, "footer.yaml: copyright is required")
	}
	if f.Commit.URL != "" && !isSafeExternalURL(f.Commit.URL) {
		*errs = append(*errs, fmt.Sprintf("footer.yaml: commit.url %q must be a valid http(s) URL", f.Commit.URL))
	}
}

func isSafeHref(raw string) bool {
	if raw == "" {
		return false
	}
	if raw == "#" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "/") {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https" || scheme == "mailto"
}

func isSafeExternalURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	return scheme == "http" || scheme == "https"
}
