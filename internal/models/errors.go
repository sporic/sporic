package models

import "errors"

var ErrInvalidCredentials = errors.New("models: invalid credentials")
var ErrRecordNotFound = errors.New("models: record not found")
