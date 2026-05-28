package bus

import (
	"errors"
	"fmt"
)

type BusError struct {
	StatusCode   int
	StatusDetail string
	Err          error
}

func (e *BusError) Error() string {
	return e.compileError().Error()
}

func (e *BusError) compileError() error {
	if e.Err != nil {
		return errors.New(fmt.Sprintf("Bus Message Status: %d - %s (error:  %v)", e.StatusCode, e.StatusDetail, e.Err))
	} else {
		return errors.New(fmt.Sprintf("Bus Message Status: %d - %s", e.StatusCode, e.StatusDetail))
	}
}

type BusErrorNotFound struct {
	BusError
}

func NotFound(otherDetail string) error {
	return &BusErrorNotFound{
		BusError: BusError{
			StatusCode:   404,
			StatusDetail: otherDetail,
			Err:          nil,
		},
	}
}

type BusErrorBadRequest struct {
	BusError
}

func BadRequest(otherDetail string, err error) error {
	return &BusErrorBadRequest{
		BusError: BusError{
			StatusCode:   400,
			StatusDetail: otherDetail,
			Err:          err,
		},
	}
}

type BusErrSystemFailure struct {
	BusError
}

func SystemFailure(otherDetail string, err error) error {
	return &BusErrSystemFailure{
		BusError: BusError{
			StatusCode:   500,
			StatusDetail: otherDetail,
			Err:          err,
		},
	}
}
