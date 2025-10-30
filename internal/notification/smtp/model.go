package smtp

type Email struct {
	Host     string
	Port     string
	To       string
	From     string
	Subject  string
	Message  string
	User     string
	Password string
}
