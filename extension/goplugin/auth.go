// Copyright (C) 2017 NTT Innovation Institute, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goplugin

import (
	"github.com/cloudwan/gohan/extension/goext"
	"github.com/cloudwan/gohan/schema"
)

// Auth is an implementation of IAuth
type Auth struct{}

// IsAdmin returns true if admin user is logged in to the admin tenant.
func (auth *Auth) IsAdminContext(context goext.Context) bool {
	if roleRaw, ok := context["role"]; ok {
		if role, ok := roleRaw.(*schema.Role); ok {
			if authRaw, ok := context["auth"]; ok {
				if auth, ok := authRaw.(schema.Authorization); ok {
					return role.Match("admin") && auth.TenantName() == "admin"
				}
				log.Warning("IsAdmin: invalid type of 'auth' field in context")
			} else {
				log.Warning("IsAdmin: missing 'auth' field in context")
			}
		} else {
			log.Warning("IsAdmin: invalid type of 'role' field in context")
		}
	} else {
		log.Warning("IsAdmin: missing 'role' field in context")
	}
	return false
}
