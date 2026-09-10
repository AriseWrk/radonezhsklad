package handler

import apperr "github.com/radonezhsklad/shared/errors"

func errBadRequest(msg string) error { return apperr.BadRequest(msg) }