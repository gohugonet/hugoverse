package localescompressed

import (
	"time"

	"github.com/gohugoio/locales"
	"github.com/gohugoio/locales/currency"
)

// The following Functions are for overriding, debugging or developing
// with a Translator Locale
// Locale returns the string value of the translator
func (ln *localen) Locale() string {
	return ln.fnLocale(ln)
}

// returns an array of cardinal plural rules associated
// with this translator
func (ln *localen) PluralsCardinal() []locales.PluralRule {
	return ln.fnPluralsCardinal(ln)
}

// returns an array of ordinal plural rules associated
// with this translator
func (ln *localen) PluralsOrdinal() []locales.PluralRule {
	return ln.fnPluralsOrdinal(ln)
}

// returns an array of range plural rules associated
// with this translator
func (ln *localen) PluralsRange() []locales.PluralRule {
	return ln.pluralsRange
}

// returns the cardinal PluralRule given 'num' and digits/precision of 'v' for locale
func (ln *localen) CardinalPluralRule(num float64, v uint64) locales.PluralRule {
	return ln.fnCardinalPluralRule(ln, num, v)
}

// returns the ordinal PluralRule given 'num' and digits/precision of 'v' for locale
func (ln *localen) OrdinalPluralRule(num float64, v uint64) locales.PluralRule {
	return ln.fnOrdinalPluralRule(ln, num, v)
}

// returns the ordinal PluralRule given 'num1', 'num2' and digits/precision of 'v1' and 'v2' for locale
func (ln *localen) RangePluralRule(num1 float64, v1 uint64, num2 float64, v2 uint64) locales.PluralRule {
	return ln.fnRangePluralRule(ln, num1, v1, num2, v2)
}

// returns the locales abbreviated month given the 'month' provided
func (ln *localen) MonthAbbreviated(month time.Month) string {
	return ln.fnMonthAbbreviated(ln, month)
}

// returns the locales abbreviated months
func (ln *localen) MonthsAbbreviated() []string {
	return ln.monthsAbbreviated
}

// returns the locales narrow month given the 'month' provided
func (ln *localen) MonthNarrow(month time.Month) string {
	return ln.monthsNarrow[month]
}

// returns the locales narrow months
func (ln *localen) MonthsNarrow() []string {
	return ln.monthsNarrow[1:]
}

// returns the locales wide month given the 'month' provided
func (ln *localen) MonthWide(month time.Month) string {
	return ln.monthsWide[month]
}

// returns the locales wide months
func (ln *localen) MonthsWide() []string {
	return ln.monthsWide[1:]
}

// returns the locales abbreviated weekday given the 'weekday' provided
func (ln *localen) WeekdayAbbreviated(weekday time.Weekday) string {
	return ln.fnWeekdayAbbreviated(ln, weekday)
}

// returns the locales abbreviated weekdays
func (ln *localen) WeekdaysAbbreviated() []string {
	return ln.daysAbbreviated
}

// returns the locales narrow weekday given the 'weekday' provided
func (ln *localen) WeekdayNarrow(weekday time.Weekday) string {
	return ln.daysNarrow[weekday]
}

// WeekdaysNarrowreturns the locales narrow weekdays
func (ln *localen) WeekdaysNarrow() []string {
	return ln.daysNarrow
}

// returns the locales short weekday given the 'weekday' provided
func (ln *localen) WeekdayShort(weekday time.Weekday) string {
	return ln.daysShort[weekday]
}

// returns the locales short weekdays
func (ln *localen) WeekdaysShort() []string {
	return ln.daysShort
}

// returns the locales wide weekday given the 'weekday' provided
func (ln *localen) WeekdayWide(weekday time.Weekday) string {
	return ln.daysWide[weekday]
}

// returns the locales wide weekdays
func (ln *localen) WeekdaysWide() []string {
	return ln.daysWide
}

// The following Functions are common Formatting functionsfor the Translator's Locale
// returns 'num' with digits/precision of 'v' for locale and handles both Whole and Real numbers based on 'v'
func (ln *localen) FmtNumber(num float64, v uint64) string {
	return ln.fnFmtNumber(ln, num, v)
}

// returns 'num' with digits/precision of 'v' for locale and handles both Whole and Real numbers based on 'v'
// NOTE: 'num' passed into FmtPercent is assumed to be in percent already
func (ln *localen) FmtPercent(num float64, v uint64) string {
	return ln.fnFmtPercent(ln, num, v)
}

// returns the currency representation of 'num' with digits/precision of 'v' for locale
func (ln *localen) FmtCurrency(num float64, v uint64, currency currency.Type) string {
	return ln.fnFmtCurrency(ln, num, v, currency)
}

// returns the currency representation of 'num' with digits/precision of 'v' for locale
// in accounting notation.
func (ln *localen) FmtAccounting(num float64, v uint64, currency currency.Type) string {
	return ln.fnFmtAccounting(ln, num, v, currency)
}

// returns the short date representation of 't' for locale
func (ln *localen) FmtDateShort(t time.Time) string {
	return ln.fnFmtDateShort(ln, t)
}

// returns the medium date representation of 't' for locale
func (ln *localen) FmtDateMedium(t time.Time) string {
	return ln.fnFmtDateMedium(ln, t)
}

//  returns the long date representation of 't' for locale
func (ln *localen) FmtDateLong(t time.Time) string {
	return ln.fnFmtDateLong(ln, t)
}

// returns the full date representation of 't' for locale
func (ln *localen) FmtDateFull(t time.Time) string {
	return ln.fnFmtDateFull(ln, t)
}

// returns the short time representation of 't' for locale
func (ln *localen) FmtTimeShort(t time.Time) string {
	return ln.fnFmtTimeShort(ln, t)
}

// returns the medium time representation of 't' for locale
func (ln *localen) FmtTimeMedium(t time.Time) string {
	return ln.fnFmtTimeMedium(ln, t)
}

// returns the long time representation of 't' for locale
func (ln *localen) FmtTimeLong(t time.Time) string {
	return ln.fnFmtTimeLong(ln, t)
}

// returns the full time representation of 't' for locale
func (ln *localen) FmtTimeFull(t time.Time) string {
	return ln.fnFmtTimeFull(ln, t)
}
