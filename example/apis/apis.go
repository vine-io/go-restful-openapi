/*
Copyright 2024 The foo Authors

This program is offered under a commercial and under the AGPL license.
For AGPL licensing, see below.

AGPL licensing:
This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package apis

import (
	restspec "github.com/vine-io/go-restful-openapi"

	"github.com/vine-io/go-restful-openapi/example/apis/internal"
)

// User is just a sample type
type User struct {
	internal.UserS
}

// Role is just a sample role type
type Role struct {
	Name string `json:"name" description:"name of the role"`
}

type UserPatch struct {
	Display string `json:"display"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

type UpdateUserInput struct {
	ID        string           `in:"path=id"`
	UID       string           `in:"query=uid;default=hello"`
	Languages []int            `in:"form=languages;default=1,2;required"`
	Cover     []*restspec.File `in:"form=cover"`
	//Payload []UserPatch  `in:"body=json"`
}
