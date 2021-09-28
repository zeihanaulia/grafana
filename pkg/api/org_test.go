package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
)

var (
	getCurrentOrgUrl        = "/api/org/"
	getOrgsUrl              = "/api/orgs/%v"
	putCurrentOrgUrl        = "/api/org/"
	putOrgsUrl              = "/api/orgs/%v"
	putCurrentOrgAddressUrl = "/api/org/address"
	putOrgsAddressUrl       = "/api/orgs/%v/address"

	updateOrgNameForm    = `{ "name": "TestOrgChanged" }`
	updateOrgAddressForm = `{ "address1": "1 test road",
	"address2": "2 test road",
	"city": "TestCity",
	"ZipCode": "TESTZIPCODE",
	"State": "TestState",
	"Country": "TestCountry" }`
)

func TestAPIEndpoint_GetCurrentOrg_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	t.Run("Viewer can view CurrentOrg", func(t *testing.T) {
		response := callAPI(server, http.MethodGet, getCurrentOrgUrl, nil, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_GetCurrentOrg_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	t.Run("AccessControl allows viewing CurrentOrg with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsRead, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodGet, getCurrentOrgUrl, nil, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("AccessControl allows viewing CurrentOrg with exact permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsRead, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodGet, getCurrentOrgUrl, nil, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("AccessControl prevents viewing CurrentOrg with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodGet, getCurrentOrgUrl, nil, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}

func TestAPIEndpoint_GetOrg_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_ADMIN, Login: testUserLogin, IsGrafanaAdmin: true}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	t.Run("Admin can view another Org", func(t *testing.T) {
		response := callAPI(server, http.MethodGet, fmt.Sprintf(getOrgsUrl, 2), nil, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_GetOrg_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	t.Run("AccessControl allows viewing another org with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsRead, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodGet, fmt.Sprintf(getOrgsUrl, 2), nil, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("AccessControl prevents viewing another org with too narrow permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsRead, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodGet, fmt.Sprintf(getOrgsUrl, 2), nil, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
	t.Run("AccessControl prevents viewing another org with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodGet, fmt.Sprintf(getOrgsUrl, 2), nil, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}

func TestAPIEndpoint_PutCurrentOrg_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_ADMIN, Login: testUserLogin}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgNameForm)

	t.Run("Admin can update current org", func(t *testing.T) {
		response := callAPI(server, http.MethodPut, putCurrentOrgUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_PutCurrentOrg_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgNameForm)
	t.Run("AccessControl allows updating current org with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodPut, putCurrentOrgUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	input = strings.NewReader(updateOrgNameForm)
	t.Run("AccessControl allows updating current org with exact permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodPut, putCurrentOrgUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("AccessControl prevents updating current org with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodPut, putCurrentOrgUrl, input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}

func TestAPIEndpoint_PutOrg_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_ADMIN, Login: testUserLogin, IsGrafanaAdmin: true}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgNameForm)

	t.Run("Admin can update another org", func(t *testing.T) {
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsUrl, 2), input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_PutOrg_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgNameForm)
	t.Run("AccessControl allows updating another org with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsUrl, 2), input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	input = strings.NewReader(updateOrgNameForm)
	t.Run("AccessControl prevents updating another org with too narrow permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsUrl, 2), input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})

	t.Run("AccessControl prevents updating another org with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsUrl, 2), input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}

func TestAPIEndpoint_PutCurrentOrgAddress_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_ADMIN, Login: testUserLogin}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgAddressForm)

	t.Run("Admin can update current org address", func(t *testing.T) {
		response := callAPI(server, http.MethodPut, putCurrentOrgAddressUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_PutCurrentOrgAddress_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgAddressForm)
	t.Run("AccessControl allows updating current org address with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodPut, putCurrentOrgAddressUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	input = strings.NewReader(updateOrgAddressForm)
	t.Run("AccessControl allows updating current org address with exact permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodPut, putCurrentOrgAddressUrl, input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	t.Run("AccessControl prevents updating current org address with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodPut, putCurrentOrgAddressUrl, input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}

func TestAPIEndpoint_PutOrgAddress_LegacyAccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_ADMIN, Login: testUserLogin, IsGrafanaAdmin: true}
	server, hs, _ := setupHTTPServer(t, false, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgAddressForm)

	t.Run("Admin can update another org address", func(t *testing.T) {
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsAddressUrl, 2), input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestAPIEndpoint_PutOrgAddress_AccessControl(t *testing.T) {
	testuser := &models.SignedInUser{UserId: testUserID, OrgId: 1, OrgRole: models.ROLE_VIEWER, Login: testUserLogin}
	server, hs, acmock := setupHTTPServer(t, true, testuser)

	// Create two orgs, to fetch another one than the logged in one
	_, err := hs.SQLStore.CreateOrgWithMember("TestOrg", testUserID)
	require.NoError(t, err)
	_, err = hs.SQLStore.CreateOrgWithMember("TestOrg2", testUserID)
	require.NoError(t, err)

	input := strings.NewReader(updateOrgAddressForm)
	t.Run("AccessControl allows updating another org address with correct permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: ScopeOrgsAll}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsAddressUrl, 2), input, t)
		assert.Equal(t, http.StatusOK, response.Code)
	})

	input = strings.NewReader(updateOrgAddressForm)
	t.Run("AccessControl prevents updating another org address with too narrow permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: ActionOrgsWrite, Scope: "orgs:1"}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsAddressUrl, 2), input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})

	t.Run("AccessControl prevents updating another org address with incorrect permissions", func(t *testing.T) {
		setAccessControlPermissions(acmock, []*accesscontrol.Permission{{Action: "orgs:invalid"}})
		response := callAPI(server, http.MethodPut, fmt.Sprintf(putOrgsAddressUrl, 2), input, t)
		assert.Equal(t, http.StatusForbidden, response.Code)
	})
}
