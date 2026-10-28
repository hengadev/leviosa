package settings

const (
	CompanyName          = "company_name"
	CompanyLogo          = "logo"
	CompanyEmail         = "support_email"
	CompanyPhone         = "telephone"
	CompanyLegalAddress  = "headquarters"
	CompanyInstagram     = "instagram_path"
	OTPDuration          = "otp_duration"
	OTPLength            = "otp_length"
	OTPMaxAttempts       = "otp_max_attempts"
	AccessTokenDuration  = "access_token_duration"
	RefreshTokenDuration = "refresh_token_duration"

	// Legal identity fields (mentions légales / GDPR compliance foundations)
	CompanyRegistrationNumber = "company_registration_number" // SIRET/SIREN
	CompanyVatNumber          = "company_vat_number"           // may be blank (franchise en base de TVA)
	CompanyHostingProvider    = "hosting_provider"              // hosting provider name + address

	// Practice page: the partner and product the public practice page shows
	// (see docs/issues/009-practice-page.md). The video URL is optional — unset
	// means the page shows its placeholder frame.
	PracticePagePartnerID = "practice_page_partner_id"
	PracticePageProductID = "practice_page_product_id"
	PracticePageVideoURL  = "practice_page_video_url"
)

// NOTE: the key that I am going to keep
// CompanyName          = "company_name"
// CompanyLogo          = "logo"
// CompanyEmail         = "support_email"
// CompanyPhone         = "telephone"
// CompanyLegalAddress  = "headquarters"
// CompanyInstagram     = "instagram_path"

// NOTE: set by me, Livio will never change that
// OTPDuration          = "otp_duration"
// OTPLength            = "otp_length"
// OTPMaxAttempts       = "otp_max_attempts"
// AccessTokenDuration  = "access_token_duration"
// RefreshTokenDuration = "refresh_token_duration"
