package airalo

import (
	"encoding/json"

	"context"
)

// RefundReason is the required, API-defined reason code for a refund request.
type RefundReason string

const (
	RefundReasonInstallationFailure    RefundReason = "INSTALLATION_FAILURE"
	RefundReasonNoCoverage             RefundReason = "NO_COVERAGE"
	RefundReasonAPNFailure             RefundReason = "APN_FAILURE"
	RefundReasonTripCancellation       RefundReason = "TRIP_CANCELLATION"
	RefundReasonIntermittentConnection RefundReason = "INTERMITTENT_CONNECTION"
	RefundReasonBlockedNetwork         RefundReason = "BLOCKED_NETWORK"
	RefundReasonChangeOfPlan           RefundReason = "CHANGE_OF_PLAN"
	RefundReasonDeletedESim            RefundReason = "DELETED_ESIM"
	RefundReasonEarlyExpiry            RefundReason = "EARLY_EXPIRY"
	RefundReasonHotspotNotWorking      RefundReason = "HOTSPOT_NOT_WORKING"
	RefundReasonIMSIChange             RefundReason = "IMSI_CHANGE"
	RefundReasonIncompatibleDevice     RefundReason = "INCOMPATIBLE_DEVICE"
	RefundReasonLockedDevice           RefundReason = "LOCKED_DEVICE"
	RefundReasonNoVoiceTextServices    RefundReason = "NO_VOICE_TEXT_SERVICES"
	RefundReasonOvercharged            RefundReason = "OVERCHARGED"
	RefundReasonSlowSpeed              RefundReason = "SLOW_SPEED"
	RefundReasonTopUpPackageFailure    RefundReason = "TOP_UP_PACKAGE_FAILURE"
	RefundReasonUnknownCharges         RefundReason = "UNKNOWN_CHARGES"
	RefundReasonWrongPurchase          RefundReason = "WRONG_PURCHASE"
	RefundReasonUnableToAccessApps     RefundReason = "UNABLE_TO_ACCESS_APPS"
	RefundReasonServiceDegradation     RefundReason = "SERVICE_DEGRADATION"
	RefundReasonQRIssuePartners        RefundReason = "QR_ISSUE_PARTNERS"
	RefundReasonOthers                 RefundReason = "OTHERS"
)

// RequestRefundParams configures POST /v2/refund.
type RequestRefundParams struct {
	// ICCIDs lists up to 5 eSIM ICCIDs to refund. Required.
	ICCIDs []string
	// Reason is required.
	Reason RefundReason
	// Notes is required when Reason is RefundReasonOthers.
	Notes string
	// Email, if set, must be a valid email address.
	Email string
}

func (p RequestRefundParams) toFormFields() map[string]string {
	fields := map[string]string{
		"reason": string(p.Reason),
	}
	if p.Notes != "" {
		fields["notes"] = p.Notes
	}
	if p.Email != "" {
		fields["email"] = p.Email
	}
	return fields
}

// RequestRefund submits a refund request for up to 5 eSIMs.
//
// The source API documentation did not include a worked response example for
// this endpoint, so no response body is decoded; a nil error indicates the
// request was accepted.
func (c *Client) RequestRefund(ctx context.Context, params RequestRefundParams) error {
	_, err := do[json.RawMessage](ctx, c, requestOptions{
		method:     "POST",
		path:       "/refund",
		formFields: params.toFormFields(),
		formArrays: map[string][]string{
			"iccids": params.ICCIDs,
		},
		authorized: true,
	})
	return err
}
