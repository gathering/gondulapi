package gondulapi

type Getter interface {
	Get(element string) error
}
type Putter interface {
	Put(element string) (error)
}
type Poster interface {
	Post() (error)
}
type Deleter interface {
	Delete(element string) (error)
}
