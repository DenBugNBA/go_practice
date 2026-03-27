package main

import (
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
)

func logClose(closer io.Closer, obj string) {
	if err := closer.Close(); err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("close %s", obj))
	}
}
