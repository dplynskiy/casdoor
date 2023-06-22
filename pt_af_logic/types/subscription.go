package types

import (
	"fmt"
	"strings"
)

type SubscriptionRole int

const (
	SubscriptionRoleUnknown SubscriptionRole = iota
	SubscriptionRoleGlobalAdmin
	SubscriptionRolePartner
	SubscriptionRoleDistributor
)

type SubscriptionStateName string

func (s SubscriptionStateName) String() string {
	return string(s)
}

const (
	SubscriptionNew           SubscriptionStateName = "New"
	SubscriptionPending       SubscriptionStateName = "Pending"
	SubscriptionPreAuthorized SubscriptionStateName = "PreAuthorized"
	SubscriptionUnauthorized  SubscriptionStateName = "Unauthorized"
	SubscriptionAuthorized    SubscriptionStateName = "Authorized"
	SubscriptionStarted       SubscriptionStateName = "Started"
	SubscriptionPreFinished   SubscriptionStateName = "PreFinished"
	SubscriptionFinished      SubscriptionStateName = "Finished"
	SubscriptionCancelled     SubscriptionStateName = "Cancelled"
)

type SubscriptionStateNames []SubscriptionStateName

func (s SubscriptionStateNames) Contains(name SubscriptionStateName) bool {
	for _, value := range s {
		if value == name {
			return true
		}
	}

	return false
}

func (s SubscriptionStateNames) String() string {
	var strs []string
	for _, state := range s {
		strs = append(strs, state.String())
	}
	return strings.Join(strs, ", ")
}

type SubscriptionFieldName string

const (
	SubscriptionFieldNameName        SubscriptionFieldName = "Name"
	SubscriptionFieldNameDisplayName SubscriptionFieldName = "Display Name"
	SubscriptionFieldNameStartDate   SubscriptionFieldName = "Start Date"
	SubscriptionFieldNameEndDate     SubscriptionFieldName = "End Date"
	SubscriptionFieldNameSubUser     SubscriptionFieldName = "Sub user"
	SubscriptionFieldNameSubPlan     SubscriptionFieldName = "Sub plan"
	SubscriptionFieldNameDiscount    SubscriptionFieldName = "Discount"
	SubscriptionFieldNameDescription SubscriptionFieldName = "Description"
)

type SubscriptionFieldNames []SubscriptionFieldName

func (s SubscriptionFieldNames) Contains(name SubscriptionFieldName) bool {
	if s == nil {
		return false
	}

	for _, value := range s {
		if value == name {
			return true
		}
	}

	return false
}

type SubscriptionFieldPermissions map[SubscriptionRole]SubscriptionFieldNames
type SubscriptionTransitions map[SubscriptionRole]SubscriptionStateNames
type SubscriptionState struct {
	FieldPermissions SubscriptionFieldPermissions
	Transitions      SubscriptionTransitions
}

var SubscriptionStateMap = map[SubscriptionStateName]SubscriptionState{
	SubscriptionNew: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameName,
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameSubUser,
				SubscriptionFieldNameSubPlan,
				SubscriptionFieldNameDiscount,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRolePartner: SubscriptionStateNames{SubscriptionPending},
		},
	},
	SubscriptionPending: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameSubPlan,
				SubscriptionFieldNameDiscount,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: nil,
	},
	SubscriptionPreAuthorized: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRolePartner: SubscriptionStateNames{SubscriptionAuthorized, SubscriptionCancelled},
		},
	},
	SubscriptionUnauthorized: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameSubPlan,
				SubscriptionFieldNameDiscount,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRolePartner: SubscriptionStateNames{SubscriptionPending, SubscriptionCancelled},
		},
	},
	SubscriptionAuthorized: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRoleDistributor: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameStartDate,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRoleDistributor: SubscriptionStateNames{SubscriptionStarted, SubscriptionCancelled},
		},
	},
	SubscriptionStarted: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRolePartner: SubscriptionStateNames{SubscriptionPreFinished},
		},
	},
	SubscriptionPreFinished: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRoleDistributor: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameEndDate,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: SubscriptionTransitions{
			SubscriptionRoleDistributor: SubscriptionStateNames{SubscriptionFinished},
		},
	},
	SubscriptionFinished: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: nil,
	},
	SubscriptionCancelled: {
		FieldPermissions: SubscriptionFieldPermissions{
			SubscriptionRolePartner: {
				SubscriptionFieldNameDisplayName,
				SubscriptionFieldNameDescription,
			},
		},
		Transitions: nil,
	},
}

func NewStateChangeForbiddenError(statusName SubscriptionStateName) error {
	return fmt.Errorf("You are not allowed to change the state to %s", statusName)
}
