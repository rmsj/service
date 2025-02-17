// Package role represents the role type in the system.
package role

import (
	"fmt"
	"slices"
)

// The set of roles that can be used.
var (
	User    = newRole("user")
	Staff   = newRole("staff")
	Support = newRole("support")
	Manager = newRole("manager")
	Admin   = newRole("admin")
)

var AllRoles = []Role{
	User,
	Staff,
	Manager,
	Support,
	Admin,
}

var AdminRoles = AllRoles

var SupportRoles = []Role{
	User,
	Staff,
	Manager,
	Support,
}

var ManagerRoles = []Role{
	User,
	Staff,
	Manager,
}

var StaffRoles = []Role{
	User,
	Staff,
}

var UserRoles = []Role{
	User,
}

// =============================================================================

// Set of known roles.
var roles = make(map[string]Role)

// Role represents a role in the system.
type Role struct {
	value string
}

func newRole(role string) Role {
	r := Role{role}
	roles[role] = r
	return r
}

// String returns the name of the role.
func (r Role) String() string {
	return r.value
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (r *Role) UnmarshalText(data []byte) error {
	role, err := Parse(string(data))
	if err != nil {
		return err
	}

	r.value = role.value
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.value), nil
}

// Equal provides support for the go-cmp package and testing.
func (r Role) Equal(r2 Role) bool {
	return r.value == r2.value
}

// =============================================================================

// Parse parses the string value and returns a role if one exists.
func Parse(value string) (Role, error) {
	role, exists := roles[value]
	if !exists {
		return Role{}, fmt.Errorf("invalid role %q", value)
	}

	return role, nil
}

// MustParse parses the string value and returns a role if one exists. If
// an error occurs the function panics.
func MustParse(value string) Role {
	role, err := Parse(value)
	if err != nil {
		panic(err)
	}

	return role
}

// ParseToString takes a collection of user roles and converts them to
// a slice of string.
func ParseToString(usrRoles []Role) []string {
	roles := make([]string, len(usrRoles))
	for i, role := range usrRoles {
		roles[i] = role.String()
	}

	return roles
}

// ParseMany takes a collection of strings and converts them to a slice
// of roles.
func ParseMany(roles []string) ([]Role, error) {
	usrRoles := make([]Role, len(roles))
	for i, roleStr := range roles {
		role, err := Parse(roleStr)
		if err != nil {
			return nil, err
		}
		usrRoles[i] = role
	}

	return usrRoles, nil
}

// HasRole checks if a user has a specific role
func HasRole(usrRoles []Role, desiredRole Role) bool {
	return slices.Contains(usrRoles, desiredRole)
}

// Set returns the set of roles for a given role
// this is defined to give an admin all the roles in the system, for example, as we do not have a hierarchical role structure
// in the authorization engine
func Set(roles []Role) []Role {
	if slices.Contains(roles, Admin) {
		return AdminRoles
	}

	if slices.Contains(roles, Support) {
		return SupportRoles
	}

	if slices.Contains(roles, Manager) {
		return ManagerRoles
	}

	if slices.Contains(roles, Staff) {
		return StaffRoles
	}

	return roles
}

// MaxRole returns the highest role in the list of roles for a given user. This is used for display purposes
func MaxRole(roles []Role) Role {

	if HasRole(roles, Admin) {
		return Admin
	}

	if HasRole(roles, Support) {
		return Support
	}

	if HasRole(roles, Manager) {
		return Manager
	}

	if HasRole(roles, Staff) {
		return Staff
	}

	return User
}
