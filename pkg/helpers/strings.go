package helpers

import (
	"fmt"
	"github.com/spf13/cast"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// UniqueStringsSorted returns a sorted slice with any duplicates removed.
// It will modify the input slice.
func UniqueStringsSorted(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	ss := sort.StringSlice(s)
	ss.Sort()
	i := 0
	for j := 1; j < len(s); j++ {
		if !ss.Less(i, j) {
			continue
		}
		i++
		s[i] = s[j]
	}

	return s[:i+1]
}

// CountWords returns the approximate word count in s.
func CountWords(s any) (int, error) {
	ss, err := cast.ToStringE(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert content to string: %w", err)
	}

	isCJKLanguage, err := IsCJKLanguage(ss)
	if err != nil {
		return 0, fmt.Errorf("failed to match regex pattern against string: %w", err)
	}

	if !isCJKLanguage {
		return len(strings.Fields(StripHTML(ss))), nil
	}

	counter := 0
	for _, word := range strings.Fields(StripHTML(ss)) {
		runeCount := utf8.RuneCountInString(word)
		if len(word) == runeCount {
			counter++
		} else {
			counter += runeCount
		}
	}

	return counter, nil
}

func IsCJKLanguage(s string) (bool, error) {
	return regexp.MatchString(`\p{Han}|\p{Hangul}|\p{Hiragana}|\p{Katakana}`, s)
}

func ReadingTime(s any) (int, error) {
	wordCount, err := CountWords(s)
	if err != nil {
		return 0, err
	}

	ss, err := cast.ToStringE(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert content to string: %w", err)
	}

	isCJKLanguage, err := IsCJKLanguage(ss)
	if err != nil {
		return 0, fmt.Errorf("failed to match CJK regex pattern against string: %w", err)
	}

	var readingTime int
	if isCJKLanguage {
		readingTime = (wordCount + 500) / 501
	} else {
		readingTime = (wordCount + 212) / 213
	}

	return readingTime, nil
}
