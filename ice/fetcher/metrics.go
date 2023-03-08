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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/iceming123/go-ice/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewRegisteredMeter("ice/fetcher/prop/announces/in", nil)
	propAnnounceOutTimer  = metrics.NewRegisteredTimer("ice/fetcher/prop/announces/out", nil)
	propAnnounceDropMeter = metrics.NewRegisteredMeter("ice/fetcher/prop/announces/drop", nil)
	propAnnounceDOSMeter  = metrics.NewRegisteredMeter("ice/fetcher/prop/announces/dos", nil)

	propBroadcastInMeter      = metrics.NewRegisteredMeter("ice/fetcher/prop/broadcasts/in", nil)
	propBroadcastOutTimer     = metrics.NewRegisteredTimer("ice/fetcher/prop/broadcasts/out", nil)
	propBroadcastDropMeter    = metrics.NewRegisteredMeter("ice/fetcher/prop/broadcasts/drop", nil)
	propBroadcastInvaildMeter = metrics.NewRegisteredMeter("ice/fetcher/prop/broadcasts/invaild", nil)
	propBroadcastDOSMeter     = metrics.NewRegisteredMeter("ice/fetcher/prop/broadcasts/dos", nil)

	propSignInvaildMeter = metrics.NewRegisteredMeter("ice/fetcher/prop/signs/invaild", nil)

	headerFetchMeter = metrics.NewRegisteredMeter("ice/fetcher/fetch/headers", nil)
	bodyFetchMeter   = metrics.NewRegisteredMeter("ice/fetcher/fetch/bodies", nil)

	headerFilterInMeter  = metrics.NewRegisteredMeter("ice/fetcher/filter/headers/in", nil)
	headerFilterOutMeter = metrics.NewRegisteredMeter("ice/fetcher/filter/headers/out", nil)
	bodyFilterInMeter    = metrics.NewRegisteredMeter("ice/fetcher/filter/bodies/in", nil)
	bodyFilterOutMeter   = metrics.NewRegisteredMeter("ice/fetcher/filter/bodies/out", nil)
)
