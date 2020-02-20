package node

import "github.com/pkg/errors"

func ErrStorageConnection(e error) error {
	return errors.Wrap(e, "ledger connection failed")
}
