// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import "net"

func LookupIpAddr(ip string) (host string) {
	res, err := net.LookupAddr(ip)
	if err != nil || len(res) == 0 {
		return ip
	}

	host = res[0]
	// Trim one trailing '.'.
	if last := len(host) - 1; last >= 0 && host[last] == '.' {
		host = host[:last]
	}
	return host
}
