package help

import "github.com/iceming123/go-ice/log"

func CheckAndPrintError(err error) {
	if err != nil {
		log.Debug("CheckAndPrintError", "error", err.Error())
	}
}
