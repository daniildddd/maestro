package postgres

import "errors"

var (
	ErrNoRows             = errors.New("no rows")
	ErrUnknown            = errors.New("unknown")
	ErrDuplicate          = errors.New("record already exists")
	ErrViolatesForeignKey = errors.New("violates foreign key")
	ErrValidation         = errors.New("check constraint violation")
	ErrDeadlock           = errors.New("deadlock detected")
)
