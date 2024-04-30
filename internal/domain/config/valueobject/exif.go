package valueobject

type ExifConfig struct {
	// Regexp matching the Exif fields you want from the (massive) set of Exif info
	// available. As we cache this info to disk, this is for performance and
	// disk space reasons more than anything.
	// If you want it all, put ".*" in this config setting.
	// Note that if neither this or ExcludeFields is set, Hugo will return a small
	// default set.
	IncludeFields string

	// Regexp matching the Exif fields you want to exclude. This may be easier to use
	// than IncludeFields above, depending on what you want.
	ExcludeFields string

	// Hugo extracts the "photo taken" date/time into .Date by default.
	// Set this to true to turn it off.
	DisableDate bool

	// Hugo extracts the "photo taken where" (GPS latitude and longitude) into
	// .Long and .Lat. Set this to true to turn it off.
	DisableLatLong bool
}
