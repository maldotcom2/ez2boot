package util

func NewHandler(version string, buildDate string) *Handler {
	return &Handler{
		Version:   version,
		BuildDate: buildDate,
	}
}
