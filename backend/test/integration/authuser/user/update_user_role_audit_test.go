package user_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/authuser/domain"
	"github.com/Leviosa-care/leviosa/backend/internal/common/contracts/identity"
	tu "github.com/Leviosa-care/leviosa/backend/internal/common/testutils"
	td "github.com/Leviosa-care/leviosa/backend/test/helpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateUserRole_WritesAuditEntry confirms issue 019's acceptance
// criterion for the authuser role-update use case: a successful role update
// writes an audit entry capturing actor, action, target, and a payload
// describing what changed.
func TestUpdateUserRole_WritesAuditEntry(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	td.ClearUsersTable(t, ctx, testPool)
	td.ClearAuditEntriesTable(t, ctx, testPool)

	accessToken := tu.SetupAdminUser(t, ctx, authCtx)

	user := td.NewTestUser(t, "audit-role@example.com", "Audit", "Target")
	user.State = domain.Active
	user.Role = identity.StandardStr
	userEncx, err := domain.ProcessUserEncx(ctx, crypto, user)
	require.NoError(t, err)
	require.NoError(t, td.InsertUserEncx(t, ctx, userEncx, testPool))

	req := td.NewUpdateUserRoleRequest(t, ctx, testServerURL, user.ID, identity.PremiumStr, accessToken)
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	entry, ok := td.GetLatestAuditEntry(t, ctx, testPool, "user", user.ID.String())
	require.True(t, ok, "expected an audit entry for the role update")

	assert.Equal(t, "user.role_updated", entry.Action)
	assert.Equal(t, "user", entry.TargetType)
	assert.Equal(t, user.ID.String(), entry.TargetID)
	assert.Equal(t, identity.AdministratorStr, entry.ActorRole)
	assert.NotEqual(t, uuid.Nil, entry.ActorID, "actor should be resolved from the session, not left nil")
	assert.Equal(t, identity.StandardStr, entry.Payload["previous_role"])
	assert.Equal(t, identity.PremiumStr, entry.Payload["new_role"])
}
