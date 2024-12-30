package language

import (
	"golang.org/x/text/language"
)

const Unknown = "Unknown"

func GetLanguageName(code string) string {
	tag, err := language.Parse(code)
	if err != nil {
		return Unknown
	}

	return tagToName[tag]
}

var tagToName = map[language.Tag]string{
	language.Afrikaans:            "Afrikaans",
	language.Amharic:              "Amharic",
	language.Arabic:               "Arabic",
	language.ModernStandardArabic: "Modern Standard Arabic",
	language.Azerbaijani:          "Azerbaijani",
	language.Bulgarian:            "Bulgarian",
	language.Bengali:              "Bengali",
	language.Catalan:              "Catalan",
	language.Czech:                "Czech",
	language.Danish:               "Danish",
	language.German:               "German",
	language.Greek:                "Greek",
	language.English:              "English",
	language.AmericanEnglish:      "American English",
	language.BritishEnglish:       "British English",
	language.Spanish:              "Spanish",
	language.EuropeanSpanish:      "European Spanish",
	language.LatinAmericanSpanish: "Latin American Spanish",
	language.Estonian:             "Estonian",
	language.Persian:              "Persian",
	language.Finnish:              "Finnish",
	language.Filipino:             "Filipino",
	language.French:               "French",
	language.CanadianFrench:       "Canadian French",
	language.Gujarati:             "Gujarati",
	language.Hebrew:               "Hebrew",
	language.Hindi:                "Hindi",
	language.Croatian:             "Croatian",
	language.Hungarian:            "Hungarian",
	language.Armenian:             "Armenian",
	language.Indonesian:           "Indonesian",
	language.Icelandic:            "Icelandic",
	language.Italian:              "Italian",
	language.Japanese:             "Japanese",
	language.Georgian:             "Georgian",
	language.Kazakh:               "Kazakh",
	language.Khmer:                "Khmer",
	language.Kannada:              "Kannada",
	language.Korean:               "Korean",
	language.Kirghiz:              "Kirghiz",
	language.Lao:                  "Lao",
	language.Lithuanian:           "Lithuanian",
	language.Latvian:              "Latvian",
	language.Macedonian:           "Macedonian",
	language.Malayalam:            "Malayalam",
	language.Mongolian:            "Mongolian",
	language.Marathi:              "Marathi",
	language.Malay:                "Malay",
	language.Burmese:              "Burmese",
	language.Nepali:               "Nepali",
	language.Dutch:                "Dutch",
	language.Norwegian:            "Norwegian",
	language.Punjabi:              "Punjabi",
	language.Polish:               "Polish",
	language.Portuguese:           "Portuguese",
	language.BrazilianPortuguese:  "Brazilian Portuguese",
	language.EuropeanPortuguese:   "European Portuguese",
	language.Romanian:             "Romanian",
	language.Russian:              "Russian",
	language.Sinhala:              "Sinhala",
	language.Slovak:               "Slovak",
	language.Slovenian:            "Slovenian",
	language.Albanian:             "Albanian",
	language.Serbian:              "Serbian",
	language.SerbianLatin:         "Serbian (Latin)",
	language.Swedish:              "Swedish",
	language.Swahili:              "Swahili",
	language.Tamil:                "Tamil",
	language.Telugu:               "Telugu",
	language.Thai:                 "Thai",
	language.Turkish:              "Turkish",
	language.Ukrainian:            "Ukrainian",
	language.Urdu:                 "Urdu",
	language.Uzbek:                "Uzbek",
	language.Vietnamese:           "Vietnamese",
	language.Chinese:              "Chinese",
	language.SimplifiedChinese:    "Simplified Chinese",
	language.TraditionalChinese:   "Traditional Chinese",
	language.Zulu:                 "Zulu",
}
