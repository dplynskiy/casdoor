package pt_af_logic

import (
	"fmt"
	"strings"

	beegocontext "github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
	af_client "github.com/casdoor/casdoor/pt_af_sdk"
	"github.com/casdoor/casdoor/util"
)

const defaultPasswordLength = 12

func CreateTenant(ctx *beegocontext.Context, subscription *object.Subscription) error {
	af := af_client.NewPtAF(afHost)

	allRoles := af.GetRoles()
	if allRoles == nil {
		return fmt.Errorf("no roles found")
	}

	customer, _ := object.GetUser(subscription.User)
	allCustomerCompanyUsers, _ := object.GetUsers(customer.Owner)

	var customerCompanyAdmin *object.User
	for _, user := range allCustomerCompanyUsers {
		if user.IsAdmin {
			customerCompanyAdmin = user
			break
		}
	}

	if customerCompanyAdmin == nil {
		return fmt.Errorf("customerCompanyAdmin doesn't exist for company: %v", customer.Owner)
	}

	loginResp, err := af.Login(af_client.LoginRequest{
		Username:    afLogin,
		Password:    afPwd,
		Fingerprint: afFingerPrint,
	})
	if err != nil {
		return fmt.Errorf("af.Login: %w", err)
	}

	af.Token = loginResp.AccessToken

	tenantName := fmt.Sprintf("%s-%s", customer.Owner, customer.Name)

	// if tenant already exists - no action required
	tenant, err := af.GetTenant(af_client.GetTenantRequest{
		Name: tenantName,
	})
	if err != nil {
		return fmt.Errorf("af.GetTenant: %w", err)
	}

	if tenant != nil {
		// tenant already exist
		return nil
	}

	tenantAdminPassword, err := generatePassword(defaultPasswordLength)
	if err != nil {
		return fmt.Errorf("generatePassword for admin: %w", err)
	}

	tenantAdminName := fmt.Sprintf("%s-%s_admin", customer.Owner, customer.Name)
	request := af_client.CreateTenantRequest{
		Name:     tenantName,
		IsActive: true,
		TrafficProcessing: af_client.TrafficProcessingRequest{
			TrafficProcessingType: "agent",
		},
		Administrator: af_client.AdministratorRequest{
			Email:    customerCompanyAdmin.Email,
			Username: tenantAdminName,
			Password: tenantAdminPassword,
		},
	}

	tenant, err = af.CreateTenant(request)
	if err != nil {
		util.LogError(ctx, err.Error())
		return fmt.Errorf("af.CreateTenant: %w", err)
	}

	if tenant != nil {
		// login from tenant admin
		token, err := af.Login(af_client.LoginRequest{
			Username:    tenantAdminName,
			Password:    tenantAdminPassword,
			Fingerprint: afFingerPrint,
		})
		if err != nil {
			return fmt.Errorf("af.Login: %w", err)
		}
		af.Token = token.AccessToken

		// create proper roles
		var serviceRole *af_client.Role
		var userRORole *af_client.Role

		for _, role := range allRoles {
			if strings.EqualFold(role.Name, "service") {
				serviceRole = &role
			}

			if strings.EqualFold(role.Name, "User RO") {
				userRORole = &role
			}
		}

		if serviceRole == nil {
			return fmt.Errorf("no service role found")
		}

		if userRORole == nil {
			return fmt.Errorf("no user RO role found")
		}

		userRoleID, _ := af.CreateRole(*userRORole)
		serviceRoleID, _ := af.CreateRole(*serviceRole)

		// create users
		userROName := fmt.Sprintf("%s_%s", customer.Name, customer.Owner)
		userROPwd, err := generatePassword(defaultPasswordLength)
		if err != nil {
			return fmt.Errorf("generatePassword: %w", err)
		}
		createUserRORequest := af_client.CreateUserRequest{
			Username:               userROName,
			Password:               userROPwd,
			Email:                  customer.Email,
			Role:                   userRoleID,
			PasswordChangeRequired: true,
			IsActive:               true,
		}

		err = af.CreateUser(createUserRORequest)
		if err != nil {
			return fmt.Errorf("af.CreateUser with user RO role: %w", err)
		}

		serviceUserName := fmt.Sprintf("%s_%s_service", customer.Name, customer.Owner)
		serviceUserPwd, err := generatePassword(defaultPasswordLength)
		if err != nil {
			return fmt.Errorf("generatePassword: %w", err)
		}
		createServiceUserRequest := af_client.CreateUserRequest{
			Username:               serviceUserName,
			Password:               serviceUserPwd,
			Email:                  customerCompanyAdmin.Email,
			Role:                   serviceRoleID,
			PasswordChangeRequired: true,
			IsActive:               true,
		}

		err = af.CreateUser(createServiceUserRequest)
		if err != nil {
			return fmt.Errorf("af.CreateUser with service role: %w", err)
		}

		// update customer properties
		if customer.Properties == nil {
			customer.Properties = make(map[string]string)
		}

		customer.Properties[af_client.PtPropPref+"Tenant ID"] = tenant.ID
		customer.Properties[af_client.PtPropPref+"ClientAccountLogin"] = userROName
		customer.Properties[af_client.PtPropPref+"ServiceAccountLogin"] = serviceUserName
		customer.Properties[af_client.PtPropPref+"tenantAdminAccountLogin"] = tenantAdminName

		affected, err := object.UpdateUser(customer.GetId(), customer, []string{"properties"}, false)
		if err != nil {
			return fmt.Errorf("object.UpdateUser: %w", err)
		}

		if !affected {
			return fmt.Errorf("object.UpdateUser didn't affected rows")
		}

		// email tenant admin info and accounts for created tenant
		err = notifyPTAFTenantCreated(&PTAFTenantCreatedMessage{
			ClientName:          customer.Name,
			ClientDisplayName:   customer.DisplayName,
			ClientURL:           fmt.Sprintf("%s/users/%s/%s", ptlmHost, customer.Owner, customer.Name),
			ServiceUserName:     serviceUserName,
			ServiceUserPwd:      serviceUserPwd,
			UserROName:          userROName,
			UserROPwd:           userROPwd,
			TenantAdminName:     tenantAdminName,
			TenantAdminPassword: tenantAdminPassword,
			PTAFLoginLink:       afHost,
		}, customerCompanyAdmin.Email)
		if err != nil {
			return fmt.Errorf("notifyPTAFTenantCreated: %w", err)
		}
	}

	return nil
}
