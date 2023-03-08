// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package fastdownloader

import (
	"github.com/iceming123/go-ice/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("ice/fastdownloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("ice/fastdownloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("ice/fastdownloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("ice/fastdownloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("ice/fastdownloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("ice/fastdownloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("ice/fastdownloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("ice/fastdownloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("ice/fastdownloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("ice/fastdownloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("ice/fastdownloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("ice/fastdownloader/receipts/timeout", nil)
)
