package goodreads

type Settings struct {
	CachePath  string `json:"cache_path,omitempty"`
	APIURL     string `json:"api_url,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
	UserID     int    `json:"user_id,omitempty"`
	PerPage    int    `json:"per_page,omitempty"`
	MaxPerPage int    `json:"max_per_page,omitempty"`
	RateLimit  int    `json:"rate_limit,omitempty"`
}

func (settings *Settings) invalid() bool {
	return settings.APIKey == "" && settings.UserID == 0
}

func DefaultSettings() *Settings {
	return &Settings{
		"./cache",
		"https://www.goodreads.com",
		"",
		0,
		50,
		200,
		1000,
	}
}

func TestSettings(cachePath, apiURL string) *Settings {
	settings := DefaultSettings()
	settings.APIURL = apiURL
	settings.APIKey = "good_test"
	settings.UserID = 1
	settings.RateLimit = 1
	settings.CachePath = cachePath
	return settings
}

func InvalidateSettings(settings *Settings) {
	settings.APIKey = ""
	settings.UserID = 0
}
