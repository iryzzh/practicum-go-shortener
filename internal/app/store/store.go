package store

type Store interface {
	URL() URLRepository
	User() UserRepository
	Ping() error
}
