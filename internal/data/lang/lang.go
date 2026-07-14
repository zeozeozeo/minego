package lang

// Translate returns the English translation for a translation key, or empty string if not found.
func Translate(key string) string {
	return translations[key]
}
