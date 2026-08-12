package helpers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	settingsKeys "github.com/Leviosa-care/leviosa/backend/internal/common/contracts/settings"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	"github.com/Leviosa-care/leviosa/backend/internal/settings/domain"
	th "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// make test-func TEST_NAME=TestSetCompanyName_WritesAuditEntry TEST_PATH=test/integration/settings/audit_test.go

// TestSetCompanyName_WritesAuditEntry confirms issue 019's acceptance
// criterion for the settings field-update use cases: a successful field
// update writes an audit entry with the correct actor, action, target (the
// settings key), and payload. Company name is used as the representative
// case — every Set* method in internal/settings/application shares the same
// auditFieldUpdate helper.
func TestSetCompanyName_WritesAuditEntry(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	th.ClearSettingsTable(t, ctx, testPool)
	th.ClearAuditEntriesTable(t, ctx, testPool)
	defer tu.ClearAuthData(t, ctx, authCtx)

	accessToken := tu.SetupAdminUser(t, ctx, authCtx)

	name := "Audited Company Name"
	request := domain.SetCompanyNameRequest{Name: name}
	req := th.NewSetCompanyNameRequest(t, ctx, testServerURL, request)
	req.AddCookie(tu.CreateAuthCookie(accessToken))

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	entry, ok := th.GetLatestAuditEntry(t, ctx, testPool, "setting", settingsKeys.CompanyName)
	require.True(t, ok, "expected an audit entry for the company-name update")

	assert.Equal(t, "settings.field_updated", entry.Action)
	assert.Equal(t, "setting", entry.TargetType)
	assert.Equal(t, settingsKeys.CompanyName, entry.TargetID)
	assert.Equal(t, identity.AdministratorStr, entry.ActorRole)
	assert.NotEqual(t, uuid.Nil, entry.ActorID)
	assert.Equal(t, settingsKeys.CompanyName, entry.Payload["field"])
	assert.Equal(t, name, entry.Payload["new_value"])
}
