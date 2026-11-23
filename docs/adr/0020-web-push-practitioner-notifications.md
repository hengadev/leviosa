# ADR-0020: Web Push for Practitioner Notifications

**Status**: Accepted
**Date**: 2026-11-23

---

## Context

The practitioner (an osteopath, who is also the administrator) works from an iPhone during the practice day. They want to learn on the phone, as it happens, that a client booked online, without opening the app or reading email. Later events (a client cancellation, a payment problem, a reminder before each session) will follow (issues 045 and 046).

Leviosa is a web app (SvelteKit PWA), not a native app. On iPhone, Web Push works since iOS 16.4, only for a web app added to the home screen, and the permission must be asked from a tap.

A push subscription's endpoint is a bearer capability: with the endpoint and the server's VAPID private key, anyone can make the device show a notification. A notification also shows on a locked screen, where anyone near the phone can read it.

## Decision

**Web Push with VAPID**, sent by the backend through a port of the notification domain (`PushNotifier`), implemented with `webpush-go`. The booking domain asks to notify the practitioner without knowing about Web Push.

- **Configuration from Infisical (ADR-0017).** `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY` and `VAPID_SUBJECT` (`mailto:contact@leviosa.care`), in the `/webpush` folder, printed whole into `backend.env`. All three or none: without them the feature is off. Nothing is sent, nothing fails, and Cabinet says notifications are unavailable.
- **Subscriptions encrypted at rest.** `notification.push_subscriptions` holds one row per device (a user may have several): user, endpoint, keys, user agent and creation date. The endpoint and keys are encrypted with encx (ADR-0015). The endpoint's encx basic hash finds a device again on unsubscribe and keeps one row per device. Deleting a user deletes their subscriptions.
- **Best effort, in the background.** Sending never fails or slows the operation that triggered it, as with the emails. A subscription the push service reports gone (404 or 410) is deleted. Errors and logs never contain an endpoint.
- **Nothing sensitive in a payload.** The payload holds a French title and body: the event, the client's first name and last-name initial ("Marie D."), the product and the time in Europe/Paris. It also holds the page to open (Aujourd'hui for a booking today, Réservations otherwise). It never holds the full name, phone, email, amounts or the practice access (the door code). The payload builder accepts only those fields, and a unit test asserts what is absent.
- **Who receives.** The practitioner of the booking concerned (its partner). A practitioner booking (booked by the practitioner themselves) never notifies.
- **The switch is per device.** Cabinet's "Notifications" row asks permission on the tap, subscribes this device and stores it; switching off unsubscribes it. The service worker shows a pushed notification and opens its page, on the app's own origin only, when tapped. The worker is built from `frontend/pwa/sw.js` because a generated Workbox worker cannot import the push handlers.

## Alternatives considered

- **SMS to the practitioner.** It costs money per message, reaches the practitioner's own number through Twilio, and puts client details in a message the practice doesn't control. Rejected.
- **A native iOS app (APNs).** An App Store app to build, review and maintain for one user. Rejected.
- **A hosted push service (OneSignal, Firebase Cloud Messaging as a relay).** A third party would see who is notified and the payloads, and would add one more account and key to hold. VAPID keys in Infisical are enough. Rejected.
- **Plain-text endpoints.** Simpler, but a database dump would let anyone push to the practitioner's phone. Rejected.

## Consequences

- A new event (cancellation, payment problem, reminder) is a payload builder in `notification/domain/push.go` and one call in the booking notification adapter.
- Changing the VAPID key pair silences every subscribed device until it is switched off and on again in Cabinet. The pair is rotated only if the private key may have leaked (`docs/deployment/README.md`).
- Notifications reach a device only while it holds a subscription. Reinstalling the home-screen app, or iOS dropping the subscription, needs the switch again; the push service then reports the old one gone and it is deleted.
- Showing more of a booking on the lock screen (the full name, the amount) needs this ADR revisited.
