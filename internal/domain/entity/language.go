package entity

type Language string

const (
	LanguageEnglish          Language = "en"
	LanguagePortugueseBrazil Language = "pt-BR"
	DefaultLanguage          Language = LanguageEnglish
)

func (language Language) IsValid() bool {
	switch language {
	case LanguageEnglish, LanguagePortugueseBrazil:
		return true
	default:
		return false
	}
}
