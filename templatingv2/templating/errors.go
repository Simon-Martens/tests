package templating

import "errors"

var InvalidPathError = errors.New("Invalid path. Must be a directory.")
var NoTemplateError = errors.New("No template found for this name")
var InvalidTemplateError = errors.New("invalid template")
var FileAccessError = errors.New("could not stat file or directory")

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
