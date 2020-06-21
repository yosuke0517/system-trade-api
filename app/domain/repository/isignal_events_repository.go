package repository

type ISignalEventsRepository interface {
	Save(string) uint
	FindOne(string) uint
}
