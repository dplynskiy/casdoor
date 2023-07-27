package pt_af_logic

import (
	"errors"
	"fmt"
	"time"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/object"
	PTAFLTypes "github.com/casdoor/casdoor/pt_af_logic/types"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/builder"
)

func GetUserRole(user *object.User) PTAFLTypes.UserRole {
	if user == nil {
		return PTAFLTypes.UserRoleUnknown
	}

	if user.IsGlobalAdmin {
		return PTAFLTypes.UserRoleGlobalAdmin
	}

	if user.IsAdmin {
		return PTAFLTypes.UserRolePartner
	}
	role, _ := object.GetRole(util.GetId(builtInOrgCode, string(PTAFLTypes.UserRoleDistributor)))
	if role != nil {
		userId := user.GetId()
		for _, roleUserId := range role.Users {
			if roleUserId == userId {
				return PTAFLTypes.UserRoleDistributor
			}
		}
	}

	return PTAFLTypes.UserRoleUnknown

}

// ValidateSubscriptionStateIsAllowed checks if user has permission to assign a new subscription state
func ValidateSubscriptionStateIsAllowed(subscriptionRole PTAFLTypes.UserRole, oldStateName, nextStateName PTAFLTypes.SubscriptionStateName) error {
	if oldStateName == nextStateName {
		return nil
	}

	oldState, ok := PTAFLTypes.SubscriptionStateMap[oldStateName]
	if !ok {
		return fmt.Errorf("incorrect old state: %s", oldStateName)
	}

	roleAvailableTransitions, ok := oldState.Transitions[subscriptionRole]
	if !ok {
		return PTAFLTypes.NewStateChangeForbiddenError(roleAvailableTransitions)
	}

	if !roleAvailableTransitions.Contains(nextStateName) {
		return PTAFLTypes.NewStateChangeForbiddenError(roleAvailableTransitions)
	}

	return nil
}

func ValidateSubscriptionRequiredFieldsIsFilled(
	userRole PTAFLTypes.UserRole,
	old, new *object.Subscription,
) error {
	if userRole == PTAFLTypes.UserRoleGlobalAdmin {
		return nil
	}

	if old.State == new.State {
		return nil
	}

	newState, ok := PTAFLTypes.SubscriptionStateMap[PTAFLTypes.SubscriptionStateName(new.State)]
	if !ok {
		return fmt.Errorf("incorrect state: %s", new.State)
	}
	requiredFields := newState.RequiredFields
	for _, requiredField := range requiredFields {
		switch requiredField {
		case PTAFLTypes.SubscriptionFieldNameSubUser:
			if new.User == "" {
				return PTAFLTypes.NewRequiredFieldNotFilledError(PTAFLTypes.SubscriptionFieldNameSubUser)
			}
		case PTAFLTypes.SubscriptionFieldNameDiscount:
			if new.Discount < 15 || new.Discount > 40 || new.Discount%5 != 0 {
				return PTAFLTypes.NewRequiredFieldNotFilledError(PTAFLTypes.SubscriptionFieldNameDiscount)
			}
		case PTAFLTypes.SubscriptionFieldNameSubPlan:
			if new.Plan == "" {
				return PTAFLTypes.NewRequiredFieldNotFilledError(PTAFLTypes.SubscriptionFieldNameSubPlan)
			}
		case PTAFLTypes.SubscriptionFieldNameStartDate:
			if new.StartDate.IsZero() {
				return PTAFLTypes.NewRequiredFieldNotFilledError(PTAFLTypes.SubscriptionFieldNameStartDate)
			}
		case PTAFLTypes.SubscriptionFieldNameEndDate:
			if new.EndDate.IsZero() {
				return PTAFLTypes.NewRequiredFieldNotFilledError(PTAFLTypes.SubscriptionFieldNameEndDate)
			}
		}
	}

	return nil
}

// ValidateSubscriptionFieldsChangeIsAllowed checks if user has permission to change fields
func ValidateSubscriptionFieldsChangeIsAllowed(
	userRole PTAFLTypes.UserRole,
	old, new *object.Subscription,
) error {
	oldState, ok := PTAFLTypes.SubscriptionStateMap[PTAFLTypes.SubscriptionStateName(old.State)]
	if !ok {
		return fmt.Errorf("incorrect state: %s", new.State)
	}

	newState, ok := PTAFLTypes.SubscriptionStateMap[PTAFLTypes.SubscriptionStateName(new.State)]
	if !ok {
		return fmt.Errorf("incorrect state: %s", new.State)
	}

	oldRoleFieldPermission := oldState.FieldPermissions[userRole]
	newRoleFieldPermission := newState.FieldPermissions[userRole]

	if old.Name != new.Name {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameName)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameName)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameName)
		}
	}

	if old.DisplayName != new.DisplayName {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDisplayName)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDisplayName)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameDisplayName)
		}
	}

	if !old.StartDate.Equal(new.StartDate) {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameStartDate)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameStartDate)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameStartDate)
		}
	}

	if !old.EndDate.Equal(new.EndDate) {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameEndDate)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameEndDate)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameEndDate)
		}
	}

	if old.User != new.User {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameSubUser)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameSubUser)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameSubUser)
		}
	}

	if old.Plan != new.Plan {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameSubPlan)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameSubPlan)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameSubPlan)
		}
	}

	if old.Discount != new.Discount {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDiscount)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDiscount)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameDiscount)
		}
	}

	if old.Description != new.Description {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDescription)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameDescription)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameDescription)
		}
	}

	if old.Comment != new.Comment {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameComment)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameComment)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameComment)
		}
	}

	if old.WasPilot != new.WasPilot {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameWasPilot)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameWasPilot)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameWasPilot)
		}
	}

	if old.PilotExpiryDate != new.PilotExpiryDate {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNamePilotExpiryDate)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNamePilotExpiryDate)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNamePilotExpiryDate)
		}
	}

	if old.Approver != new.Approver {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameApprover)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameApprover)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameApprover)
		}
	}

	if old.ApproveTime != new.ApproveTime {
		oldContains := oldRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameApproveTime)
		newContains := newRoleFieldPermission.Contains(PTAFLTypes.SubscriptionFieldNameApproveTime)
		if !oldContains && !newContains {
			return PTAFLTypes.NewForbiddenFieldChangeError(PTAFLTypes.SubscriptionFieldNameApproveTime)
		}
	}

	return nil
}

func ValidateSubscriptionByState(user *object.User, subscription *object.Subscription, old *object.Subscription) error {
	if old.State == subscription.State {
		return nil
	}

	state := PTAFLTypes.SubscriptionStateName(subscription.State)
	switch state {
	case PTAFLTypes.SubscriptionPending:
		filter := builder.Neq{"state": []string{
			PTAFLTypes.SubscriptionNew.String(),
			PTAFLTypes.SubscriptionCancelled.String(),
			PTAFLTypes.SubscriptionPreFinished.String(),
			PTAFLTypes.SubscriptionFinished.String(),
		}}

		if subscription.User == "" {
			//should never happen - user is required field for pending state
			return errors.New("empty user")
		}

		userSubscriptionsCount, err := object.GetSubscriptionCount(subscription.Owner, "subscription.user", subscription.User, filter)
		if err != nil {
			return fmt.Errorf("object.GetSubscriptionCount: %w", err)
		}

		if userSubscriptionsCount > 0 {
			return errors.New(i18n.Translate(ptlmLanguage, "subscription:Customer has active subscriptions"))
		}

	case PTAFLTypes.SubscriptionPilot:
		filter := builder.And(builder.Eq{
			"subscription.user": subscription.User,
		},
			builder.Neq{
				"state": []string{
					PTAFLTypes.SubscriptionNew.String(),
					PTAFLTypes.SubscriptionCancelled.String(),
					PTAFLTypes.SubscriptionFinished.String(),
				}}.Or(builder.And(
				builder.Eq{
					"state":     PTAFLTypes.SubscriptionCancelled.String(),
					"was_pilot": true},
				builder.Expr("TO_DATE(approve_time,'YYYY-MM-DD')>TO_DATE(?,'YYYY-MM-DD')",
					time.Now().AddDate(0, -3, 0).Format("2006-01-02")),
			)).Or(builder.And(
				builder.Eq{
					"state": PTAFLTypes.SubscriptionFinished.String(),
				},
				builder.Gt{
					"end_date": time.Now().Truncate(24*time.Hour).AddDate(0, -1, 0),
				},
			)))

		if subscription.User == "" {
			//should never happen - user is required field for pending state
			return errors.New("empty user")
		}

		userSubscriptionsCount, err := object.GetSubscriptionCount(subscription.Owner, "", "", filter)
		if err != nil {
			return fmt.Errorf("object.GetSubscriptionCount(customerLimit): %w", err)
		}

		if userSubscriptionsCount > 0 {
			return errors.New(i18n.Translate(ptlmLanguage, "subscription:Customer doesn't meet the requirements for pilot"))
		}

		filterPilotLimit := builder.Eq{
			"was_pilot": true,
			"state": []string{
				PTAFLTypes.SubscriptionPilot.String(),
				PTAFLTypes.SubscriptionPending.String(),
				PTAFLTypes.SubscriptionUnauthorized.String(),
				PTAFLTypes.SubscriptionPreAuthorized.String(),
				PTAFLTypes.SubscriptionPilotExpired.String(),
			},
		}

		organization, err := object.GetOrganization(util.GetId("admin", subscription.Owner))
		if err != nil {
			return fmt.Errorf("object.GetOrganization: %w", err)
		}

		partnerPilotSubscriptionsCount, err := object.GetSubscriptionCount(subscription.Owner, "", "", filterPilotLimit)
		if err != nil {
			return fmt.Errorf("object.GetSubscriptionCount(partnerLimit): %w", err)
		}

		if uint(partnerPilotSubscriptionsCount) >= organization.PilotLimit {
			return errors.New(i18n.Translate(ptlmLanguage, "subscription:Pilot Limit exceeded"))
		}
	}

	return nil
}

func FillSubscriptionByState(user *object.User, subscription *object.Subscription, old *object.Subscription) error {
	if old.State == subscription.State {
		return nil
	}

	subscription.Approver = user.GetId()
	subscription.ApproveTime = time.Now().Format("2006-01-02T15:04:05Z07:00")

	state := PTAFLTypes.SubscriptionStateName(subscription.State)
	switch state {
	case PTAFLTypes.SubscriptionPilot:

		subscription.WasPilot = true
		mskLoc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			return fmt.Errorf("time.LoadLocation: %w", err)
		}
		expiryDate := time.Now().In(mskLoc).Truncate(24*time.Hour).AddDate(0, 1, 0)
		subscription.PilotExpiryDate = &expiryDate
	}

	return nil
}

func ValidateSubscriptionUpdate(user *object.User, subscription *object.Subscription, old *object.Subscription) error {
	subscriptionRole := GetUserRole(user)

	if subscriptionRole == PTAFLTypes.UserRoleGlobalAdmin {
		return nil
	}

	oldStateName := PTAFLTypes.SubscriptionStateName(old.State)
	newStateName := PTAFLTypes.SubscriptionStateName(subscription.State)

	err := ValidateSubscriptionStateIsAllowed(subscriptionRole, oldStateName, newStateName)
	if err != nil {
		return err
	}

	err = ValidateSubscriptionFieldsChangeIsAllowed(subscriptionRole, old, subscription)
	if err != nil {
		return err
	}

	err = ValidateSubscriptionRequiredFieldsIsFilled(subscriptionRole, old, subscription)
	if err != nil {
		return err
	}

	// additional checks for some states
	err = ValidateSubscriptionByState(user, subscription, old)
	if err != nil {
		return err
	}

	return nil
}

func ProcessSubscriptionUpdatePostActions(ctx *context.Context, user *object.User, subscription, old *object.Subscription) {
	stateChanged := old.State != subscription.State

	err := NotifySubscriptionUpdated(ctx, user, subscription, old)
	if err != nil {
		util.LogError(ctx, fmt.Errorf("NotifySubscriptionUpdated: %w", err).Error())
	}

	if !stateChanged {
		return
	}

	switch PTAFLTypes.SubscriptionStateName(subscription.State) {
	case PTAFLTypes.SubscriptionStarted, PTAFLTypes.SubscriptionPilot:
		// create or enable tenant at pt af
		err := CreateOrEnableTenant(ctx, subscription)
		if err != nil {
			util.LogError(ctx, fmt.Errorf("CreateTenant: %w", err).Error())
		}
	case PTAFLTypes.SubscriptionPilotExpired:
		// disable Tenant
		err := DisableTenant(ctx, subscription)
		if err != nil {
			util.LogError(ctx, fmt.Errorf("DisableTenant: %w", err).Error())
		}
	case PTAFLTypes.SubscriptionCancelled:
		if !subscription.WasPilot {
			return
		}
		// disable Tenant
		err := DisableTenant(ctx, subscription)
		if err != nil {
			util.LogError(ctx, fmt.Errorf("DisableTenant: %w", err).Error())
		}
	}
}

func GetSubscriptionFilter(user *object.User) builder.Cond {
	userRole := GetUserRole(user)
	if userRole == PTAFLTypes.UserRoleDistributor {
		return builder.Eq{"state": []string{
			PTAFLTypes.SubscriptionAuthorized.String(),
			PTAFLTypes.SubscriptionStarted.String(),
			PTAFLTypes.SubscriptionPreFinished.String(),
			PTAFLTypes.SubscriptionFinished.String(),
		}}.Or(builder.Eq{"state": PTAFLTypes.SubscriptionCancelled.String(), "approver": user.GetId()})
	}

	return nil
}

func GetAvailableTransitions(user *object.User, subscription *object.Subscription) ([]PTAFLTypes.SubscriptionStateName, error) {
	subscriptionRole := GetUserRole(user)

	subscriptionState := PTAFLTypes.SubscriptionStateName(subscription.State)
	state, ok := PTAFLTypes.SubscriptionStateMap[subscriptionState]
	if !ok {
		return nil, fmt.Errorf("incorrect state: %s", subscriptionState)
	}

	roleAvailableTransitions, _ := state.Transitions[subscriptionRole]
	roleAvailableTransitions = append([]PTAFLTypes.SubscriptionStateName{subscriptionState}, roleAvailableTransitions...)

	return roleAvailableTransitions, nil
}
