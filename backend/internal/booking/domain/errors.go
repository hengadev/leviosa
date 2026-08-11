package domain

import "errors"

// Building errors
var (
	ErrInvalidBuildingName    = errors.New("building name cannot be empty")
	ErrInvalidBuildingAddress = errors.New("building address cannot be empty")
	ErrInvalidBuildingCity    = errors.New("building city cannot be empty")
	ErrInvalidBuildingCountry = errors.New("building country cannot be empty")
	ErrInvalidBuildingID      = errors.New("invalid building ID")
)

// Room errors
var (
	ErrInvalidRoomName     = errors.New("room name cannot be empty")
	ErrInvalidRoomCapacity = errors.New("room capacity must be greater than 0")
	ErrInvalidRoomID       = errors.New("invalid room ID")
)

// Room allocation errors
var (
	ErrInvalidPartnerID                   = errors.New("invalid partner ID")
	ErrInvalidAllocationStartDate         = errors.New("allocation start date cannot be empty for dedicated allocations")
	ErrInvalidAllocationEndDate           = errors.New("allocation end date must be after start date")
	ErrCannotUpdateSharedAllocationPeriod = errors.New("cannot update time period for shared allocations")
)

// Availability errors
var (
	ErrInvalidTimeSlot                = errors.New("end time must be after start time")
	ErrCannotCreatePastAvailability   = errors.New("cannot create availability in the past")
	ErrInvalidAvailabilityCapacity    = errors.New("availability capacity must be greater than 0")
	ErrCannotUpdateToPastTime         = errors.New("cannot update availability to past time")
	ErrCannotUpdateBookedAvailability = errors.New("cannot update booked availability")
	ErrAvailabilityNotAvailable       = errors.New("availability is not available for booking")
	ErrInvalidRecurrenceType          = errors.New("invalid recurrence type")
	ErrInvalidRecurrenceInterval      = errors.New("recurrence interval must be greater than 0")
	ErrMissingWeeklyDays              = errors.New("weekly recurrence requires days of week")
	ErrInvalidWeekDay                 = errors.New("week day must be between 0 and 6")
	ErrInvalidAvailabilityID          = errors.New("invalid availability ID")
)

// Booking errors
var (
	ErrInvalidClientID                 = errors.New("invalid client ID")
	ErrAmbiguousBookingIdentity        = errors.New("booking must have either a client ID or guest fields, not both")
	ErrInvalidBookingPrice             = errors.New("booking price cannot be negative")
	ErrCannotMarkRefundedAsPaid        = errors.New("cannot mark refunded payment as paid")
	ErrCannotRefundUnpaidBooking       = errors.New("cannot refund unpaid booking")
	ErrCannotCancelCompletedBooking    = errors.New("cannot cancel completed booking")
	ErrBookingAlreadyCancelled         = errors.New("booking is already cancelled")
	ErrCannotCompleteBooking           = errors.New("cannot complete cancelled booking")
	ErrBookingAlreadyCompleted         = errors.New("booking is already completed")
	ErrCannotMarkCancelledAsNoShow     = errors.New("cannot mark cancelled booking as no-show")
	ErrCannotMarkCompletedAsNoShow     = errors.New("cannot mark completed booking as no-show")
	ErrCannotDisputeUnpaidBooking      = errors.New("cannot dispute a booking that is not paid")
	ErrCannotResolveNonDisputedBooking = errors.New("cannot resolve dispute for a booking that is not disputed")
)

// Booking token errors
var (
	ErrInvalidBookingToken = errors.New("invalid booking token")
	ErrBookingTokenExpired = errors.New("booking token has expired")
)

// Review errors
var (
	ErrInvalidReviewBookingID = errors.New("invalid booking ID for review")
	ErrInvalidReviewRating    = errors.New("review rating must be between 1 and 5")
	ErrReviewCommentTooLong   = errors.New("review comment exceeds maximum length")
	ErrReviewAlreadyHidden    = errors.New("review is already hidden")
	ErrReviewNotHidden        = errors.New("review is not hidden")
	ErrBookingNotCompleted    = errors.New("booking must be completed before it can be reviewed")
	ErrReviewNotBookingOwner  = errors.New("only the booking's client may leave a review for it")
	ErrReviewAlreadyExists    = errors.New("a review already exists for this booking")
)

// Waitlist entry errors
var (
	ErrInvalidWaitlistProductID      = errors.New("invalid product ID")
	ErrInvalidWaitlistDateRange      = errors.New("desired range end must be after desired range start")
	ErrCannotJoinWaitlistInPast      = errors.New("cannot join waitlist for a date range that has already passed")
	ErrInvalidWaitlistIdentity       = errors.New("waitlist entry must have either a client ID or guest fields")
	ErrAmbiguousWaitlistIdentity     = errors.New("waitlist entry must have either a client ID or guest fields, not both")
	ErrWaitlistEntryNotWaiting       = errors.New("waitlist entry is not in the waiting state")
	ErrWaitlistEntryNotNotified      = errors.New("waitlist entry is not in the notified state")
	ErrWaitlistEntryAlreadyCancelled = errors.New("waitlist entry is already cancelled")
	ErrCannotCancelWaitlistEntry     = errors.New("cannot cancel a waitlist entry that has already been claimed or expired")
)
