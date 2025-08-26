package postgres

type Settings struct {
	User string
	Pass string
	Host string
	Port string
	Name string
	SslMode string
	Reload bool
}