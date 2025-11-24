package handlers

// Request types for better type safety
type addFeedRequest struct {
	URL      string `json:"url"`
	Category string `json:"category"`
	Title    string `json:"title"`
}

type updateFeedRequest struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

type translateRequest struct {
	ArticleID  int64  `json:"article_id"`
	Title      string `json:"title"`
	TargetLang string `json:"target_language"`
}

type settingsRequest struct {
	UpdateInterval      string `json:"update_interval"`
	TranslationEnabled  string `json:"translation_enabled"`
	TargetLanguage      string `json:"target_language"`
	TranslationProvider string `json:"translation_provider"`
	DeepLAPIKey         string `json:"deepl_api_key"`
	AutoCleanupEnabled  string `json:"auto_cleanup_enabled"`
	MaxCacheSizeMB      string `json:"max_cache_size_mb"`
	MaxArticleAgeDays   string `json:"max_article_age_days"`
	Language            string `json:"language"`
	Theme               string `json:"theme"`
	ShowHiddenArticles  string `json:"show_hidden_articles"`
	StartupOnBoot       string `json:"startup_on_boot"`
}

type downloadUpdateRequest struct {
	DownloadURL string `json:"download_url"`
	AssetName   string `json:"asset_name"`
}

type installUpdateRequest struct {
	FilePath string `json:"file_path"`
}

// Response helper types
type errorResponse struct {
	Error string `json:"error"`
}

type successResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}
