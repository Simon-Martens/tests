package templating

type FSError[T error] struct {
	File string
	Err  T
}

func NewError[T error](t T, file string) FSError[T] {
	return FSError[T]{File: file, Err: t}
}

func (e FSError[T]) Error() string {
	return e.Err.Error() + ": " + e.File
}
