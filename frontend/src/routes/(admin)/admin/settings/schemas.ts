import { type } from 'arktype'
import type { Infer } from 'sveltekit-superforms'

// Company Settings Schemas
export const companyNameSchema = type({
    name: "string"
})

export const companyEmailSchema = type({
    email: "string"
})

export const companyPhoneSchema = type({
    telephone: "string"
})

export const companyAddressSchema = type({
    address: "string"
})

export const companyInstagramSchema = type({
    instagram: "string"
})

// OTP Settings Schemas
export const otpDurationSchema = type({
    duration: "number"
})

export const otpLengthSchema = type({
    length: "number"
})

export const otpMaxAttemptsSchema = type({
    max_attempts: "number"
})

// Token Settings Schemas
export const accessTokenDurationSchema = type({
    duration: "number"
})

export const refreshTokenDurationSchema = type({
    duration: "number"
})

// Practice Page Settings Schema (page publique du cabinet)
export const practicePageSchema = type({
    partner_id: "string.uuid",
    product_id: "string.uuid",
    video_url: "string",
})

// Booking Policy Settings Schema (slot engine): minimum notice in hours,
// booking horizon in weeks, payment hold in minutes (ADR-0018) — the same
// bounds the backend enforces.
export const bookingPolicySchema = type({
    minimum_notice_hours: "0 <= number.integer <= 168",
    horizon_weeks: "1 <= number.integer <= 52",
    payment_hold_minutes: "5 <= number.integer <= 60",
})

// Types
export type CompanyName = Infer<typeof companyNameSchema>
export type CompanyEmail = Infer<typeof companyEmailSchema>
export type CompanyPhone = Infer<typeof companyPhoneSchema>
export type CompanyAddress = Infer<typeof companyAddressSchema>
export type CompanyInstagram = Infer<typeof companyInstagramSchema>
export type PracticePage = Infer<typeof practicePageSchema>
export type BookingPolicy = Infer<typeof bookingPolicySchema>
export type OtpDuration = Infer<typeof otpDurationSchema>
export type OtpLength = Infer<typeof otpLengthSchema>
export type OtpMaxAttempts = Infer<typeof otpMaxAttemptsSchema>
export type AccessTokenDuration = Infer<typeof accessTokenDurationSchema>
export type RefreshTokenDuration = Infer<typeof refreshTokenDurationSchema>

// Defaults
export const companyNameDefaults: CompanyName = {
    name: "",
}

export const companyEmailDefaults: CompanyEmail = {
    email: "",
}

export const companyPhoneDefaults: CompanyPhone = {
    telephone: "",
}

export const companyAddressDefaults: CompanyAddress = {
    address: "",
}

export const companyInstagramDefaults: CompanyInstagram = {
    instagram: "",
}

export const practicePageDefaults: PracticePage = {
    partner_id: "",
    product_id: "",
    video_url: "",
}

export const bookingPolicyDefaults: BookingPolicy = {
    minimum_notice_hours: 12,
    horizon_weeks: 8,
    payment_hold_minutes: 15,
}

export const otpDurationDefaults: OtpDuration = {
    duration: 300,
}

export const otpLengthDefaults: OtpLength = {
    length: 6,
}

export const otpMaxAttemptsDefaults: OtpMaxAttempts = {
    max_attempts: 5,
}

export const accessTokenDurationDefaults: AccessTokenDuration = {
    duration: 15,
}

export const refreshTokenDurationDefaults: RefreshTokenDuration = {
    duration: 168,
}
