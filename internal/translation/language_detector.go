package translation

import (
	"strings"
	"sync"

	"github.com/pemistahl/lingua-go"
)

// LanguageDetector handles language detection using Lingua
type LanguageDetector struct {
	detector lingua.LanguageDetector
	once     sync.Once
}

// languageDetectorInstance is the singleton instance
var (
	languageDetectorInstance *LanguageDetector
	languageDetectorOnce     sync.Once
)

// Supported languages for detection
var supportedLanguages = []lingua.Language{
	lingua.English,
	lingua.Chinese,
	lingua.Japanese,
	lingua.Korean,
	lingua.Spanish,
	lingua.French,
	lingua.German,
	lingua.Portuguese,
	lingua.Russian,
	lingua.Italian,
	lingua.Arabic,
	lingua.Dutch,
	lingua.Polish,
	lingua.Turkish,
	lingua.Vietnamese,
	lingua.Thai,
	lingua.Indonesian,
	lingua.Hindi,
}

// GetLanguageDetector returns the singleton language detector instance
func GetLanguageDetector() *LanguageDetector {
	languageDetectorOnce.Do(func() {
		languageDetectorInstance = &LanguageDetector{}
		languageDetectorInstance.once.Do(func() {
			// Build detector with low accuracy mode for better performance on short texts
			languageDetectorInstance.detector = lingua.NewLanguageDetectorBuilder().
				FromLanguages(supportedLanguages...).
				WithLowAccuracyMode().
				Build()
		})
	})
	return languageDetectorInstance
}

// DetectLanguage detects the language of the given text
// Returns the ISO 639-1 language code (e.g., "en", "zh", "ja")
// Returns empty string if detection fails or confidence is too low
func (ld *LanguageDetector) DetectLanguage(text string) string {
	if text == "" {
		return ""
	}

	// Clean text for better detection
	text = strings.TrimSpace(text)
	if len(text) < 3 {
		return ""
	}

	// Remove HTML tags if present
	cleanText := removeHTMLTags(text)
	textForDetection := text

	// Only use cleaned text if it's significantly different and has enough content
	if len(cleanText) > 10 && len(cleanText) < len(text) {
		textForDetection = cleanText
	}

	// Detect language
	language, exists := ld.detector.DetectLanguageOf(textForDetection)
	if !exists {
		return ""
	}

	detectedCode := linguaToISOCode(language)
	return detectedCode
}

// ShouldTranslate determines if translation is needed based on language detection
// Returns true if:
// - Language detection fails (fallback to translation for safety)
// - Detected language differs from target language
// Returns false if text is already in target language
func (ld *LanguageDetector) ShouldTranslate(text, targetLang string) bool {
	detectedLang := ld.DetectLanguage(text)

	// If detection failed, assume translation is needed (fallback behavior)
	if detectedLang == "" {
		return true
	}

	// Normalize language codes for comparison
	detectedLang = normalizeLangCode(detectedLang)
	targetLang = normalizeLangCode(targetLang)

	// Check if already in target language
	return detectedLang != targetLang
}

// linguaToISOCode converts Lingua Language enum to ISO 639-1 code
func linguaToISOCode(language lingua.Language) string {
	langMap := map[lingua.Language]string{
		lingua.English:    "en",
		lingua.Chinese:    "zh",
		lingua.Japanese:   "ja",
		lingua.Korean:     "ko",
		lingua.Spanish:    "es",
		lingua.French:     "fr",
		lingua.German:     "de",
		lingua.Portuguese: "pt",
		lingua.Russian:    "ru",
		lingua.Italian:    "it",
		lingua.Arabic:     "ar",
		lingua.Dutch:      "nl",
		lingua.Polish:     "pl",
		lingua.Turkish:    "tr",
		lingua.Vietnamese: "vi",
		lingua.Thai:       "th",
		lingua.Indonesian: "id",
		lingua.Hindi:      "hi",
	}

	if code, ok := langMap[language]; ok {
		return code
	}
	return ""
}

// normalizeLangCode normalizes language codes (e.g., "zh-CN" -> "zh", "en-US" -> "en")
func normalizeLangCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if len(code) > 2 {
		code = code[:2]
	}
	return code
}

// removeHTMLTags removes HTML tags from text for better language detection
func removeHTMLTags(text string) string {
	// Simple HTML tag removal
	result := ""
	inTag := false
	for _, r := range text {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result += string(r)
		}
	}
	return strings.TrimSpace(result)
}
