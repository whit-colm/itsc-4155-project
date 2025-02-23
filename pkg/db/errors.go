package db

type dbError string

func (e dbError) Error() string {
	return string(e)
}

const (
	ConnectionNotEstablished dbError = "Database connection has not been established"
	ConnectionNotCreated     dbError = "Unable to create connection pool"
)
