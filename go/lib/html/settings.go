package html

// Settings represents the settings of the HTML templates
type Settings struct {
	TemplatePath string `json:"template_path,omitempty"`
	TemplateExt  string `json:"template_ext,omitempty"`
	LayoutName   string `json:"layoutName,omitempty"`
	WebsiteTitle string `json:"website_title,omitempty"`
}

// DefaultSettings is the default settings of the HTML templates
func DefaultSettings() *Settings {
	return &Settings{
		"./go/content/templates",
		".gohtml",
		"layout",
		"Your Website Title",
	}
}
