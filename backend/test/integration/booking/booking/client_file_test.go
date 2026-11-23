package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Leviosa-care/leviosa/backend/internal/booking/domain"
	tsetup "github.com/Leviosa-care/leviosa/backend/test/helpers/booking"
	tavail "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/availability"
	tbooking "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/booking"
	tb "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/building"
	tr "github.com/Leviosa-care/leviosa/backend/test/helpers/booking/room"
	tcatalog "github.com/Leviosa-care/leviosa/backend/test/helpers/catalog"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The client file (issue 042): GET /partners/bookings/{id}/client returns the
// contact details of the booking's client and their sessions with the
// practitioner; GET /partners/clients/{partnerId}?q= searches them. Nothing
// is stored and no notes appear anywhere (ADR-0019).
//
// make test-func TEST_NAME=TestClientFile TEST_PATH=test/integration/booking/booking/client_file_test.go

// clientFileFixture clears the tables, then sets up a partner with two paid
// guest bookings of the same client: one past, one upcoming. It returns the
// partner's token and id and the two booking ids.
func clientFileFixture(t *testing.T, ctx context.Context, email string, guestPhone string) (partnerToken string, partnerID, upcomingID, cancelledID uuid.UUID) {
	t.Helper()

	tbooking.ClearBookingsTable(t, ctx, testPool)
	tavail.ClearAvailabilityTable(t, ctx, testPool)
	tr.ClearRoomsTable(t, ctx, testPool)
	tb.ClearBuildingsTable(t, ctx, testPool)
	tcatalog.ClearProductsTable(t, ctx, testPool)

	categoryID := tcatalog.CreateTestCategory(t, ctx, testPool)
	products := tcatalog.CreateDefaultTestProducts(t, ctx, testPool, categoryID)

	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))

	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))

	partnerToken, partnerID = tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, email, room.ID, testPool, authCtx.Redis, crypto)

	client := &http.Client{Timeout: 10 * time.Second}
	now := time.Now()
	book := func(offset time.Duration) uuid.UUID {
		start := now.Add(offset).Truncate(time.Hour)
		availability := tavail.NewTestAvailability(t, partnerID, room.ID, start, start.Add(2*time.Hour), 5000)
		availabilityEncx, err := domain.ProcessAvailabilityEncx(ctx, crypto, availability)
		require.NoError(t, err)
		tavail.InsertAvailabilityEncx(t, ctx, availabilityEncx, testPool)

		requestBody := map[string]interface{}{
			"availability_id":  availability.ID.String(),
			"product_id":       products[0].ID.String(),
			"slot_start_time":  start.Format(time.RFC3339),
			"guest_first_name": "Jean",
			"guest_last_name":  "Dupont",
			"guest_email":      "jean.dupont@example.com",
			"guest_phone":      guestPhone,
		}
		bodyBytes, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServerURL+"/bookings", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var created domain.BookingResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		confirmBookingPayment(t, ctx, created.ID)
		return created.ID
	}

	upcomingID = book(24 * time.Hour)
	cancelledID = book(48 * time.Hour)
	cancelAsPractitioner(t, client, cancelledID, partnerToken)
	return partnerToken, partnerID, upcomingID, cancelledID
}

// otherPractitioner sets up a second practitioner with their own room.
func otherPractitioner(t *testing.T, ctx context.Context, email string) (token string, id uuid.UUID) {
	t.Helper()
	building := tb.NewTestBuilding(t)
	buildingEncx, err := domain.ProcessBuildingEncx(ctx, crypto, building)
	require.NoError(t, err)
	require.NoError(t, tb.InsertBuildingEncx(t, ctx, testPool, buildingEncx))
	room := tr.NewTestRoomWithBuilding(t, building.ID)
	roomEncx, err := domain.ProcessRoomEncx(ctx, crypto, room)
	require.NoError(t, err)
	require.NoError(t, tr.InsertRoomEncx(t, ctx, testPool, roomEncx))
	return tsetup.SetupAuthenticatedPartnerWithAllocation(t, ctx, email, room.ID, testPool, authCtx.Redis, crypto)
}

func readAll(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return body
}

func TestClientFile(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("shows the contact and both session lists", func(t *testing.T) {
		partnerToken, _, upcomingID, pastID := clientFileFixture(t, ctx, "cffile1@test.com", "0612345678")

		req, err := http.NewRequest("GET", testServerURL+"/partners/bookings/"+upcomingID.String()+"/client", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body := readAll(t, resp)
		var file domain.ClientFileResponse
		require.NoError(t, json.Unmarshal(body, &file))
		assert.Equal(t, "Jean Dupont", file.Name)
		assert.Equal(t, "0612345678", file.Phone)
		assert.Equal(t, "jean.dupont@example.com", file.Email)

		// Cancelled sessions stay listed and marked; both are future slots
		// here, so both sit in upcoming (soonest first: the confirmed one).
		require.Len(t, file.Upcoming, 2)
		assert.Equal(t, upcomingID, file.Upcoming[0].ID)
		assert.Equal(t, domain.BookingStatusConfirmed, file.Upcoming[0].Status)
		assert.Equal(t, pastID, file.Upcoming[1].ID)
		assert.Equal(t, domain.BookingStatusCancelled, file.Upcoming[1].Status)
		assert.Empty(t, file.Past)

		// The same file is reachable from the other booking: same client.
		req2, err := http.NewRequest("GET", testServerURL+"/partners/bookings/"+pastID.String()+"/client", nil)
		require.NoError(t, err)
		req2.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		resp2, err := client.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		// ADR-0019: no notes anywhere in the client file
		var raw map[string]any
		require.NoError(t, json.Unmarshal(body, &raw))
		assert.NotContains(t, raw, "client_notes")
		assert.NotContains(t, raw, "partner_notes")
	})

	t.Run("another practitioner's client is not reachable", func(t *testing.T) {
		partnerToken, _, upcomingID, _ := clientFileFixture(t, ctx, "cffile2@test.com", "0687654321")
		otherToken, _ := otherPractitioner(t, ctx, "other-cffile-"+uuid.New().String()+"@test.com")

		req, err := http.NewRequest("GET", testServerURL+"/partners/bookings/"+upcomingID.String()+"/client", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: otherToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// The owner still sees it.
		req2, err := http.NewRequest("GET", testServerURL+"/partners/bookings/"+upcomingID.String()+"/client", nil)
		require.NoError(t, err)
		req2.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		resp2, err := client.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusOK, resp2.StatusCode)
	})
}

func TestSearchPartnerClients(t *testing.T) {
	ctx := context.Background()
	client := &http.Client{Timeout: 10 * time.Second}

	t.Run("matches a name or phone fragment", func(t *testing.T) {
		partnerToken, partnerID, _, _ := clientFileFixture(t, ctx, "cffile3@test.com", "0611223344")

		for _, q := range []string{"dupont", "6112"} {
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/partners/clients/%s?q=%s", testServerURL, partnerID, q), nil)
			require.NoError(t, err)
			req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
			resp, err := client.Do(req)
			require.NoError(t, err)
			body := readAll(t, resp)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			var clients []domain.ClientSummary
			require.NoError(t, json.Unmarshal(body, &clients))
			require.Len(t, clients, 1)
			assert.Equal(t, "Jean Dupont", clients[0].Name)
			assert.Equal(t, "0611223344", clients[0].Phone)
			assert.Equal(t, "jean.dupont@example.com", clients[0].Email)
		}

		// A fragment matching nothing returns an empty list.
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/partners/clients/%s?q=zzz", testServerURL, partnerID), nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: partnerToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var clients []domain.ClientSummary
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&clients))
		assert.Empty(t, clients)
	})

	t.Run("another practitioner's clients are not searchable", func(t *testing.T) {
		_, _, _, _ = clientFileFixture(t, ctx, "cffile4@test.com", "0699887766")
		otherToken, otherID := otherPractitioner(t, ctx, "other-cffile-"+uuid.New().String()+"@test.com")

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/partners/clients/%s?q=dupont", testServerURL, otherID), nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "leviosa_access_token", Value: otherToken})
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var clients []domain.ClientSummary
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&clients))
		assert.Empty(t, clients, "each practitioner only searches their own clients")
	})
}
