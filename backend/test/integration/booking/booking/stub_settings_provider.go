package booking

import "context"

// stubSettingsProvider is a minimal notificationPorts.SettingsProvider used to
// satisfy BookingNotificationAdapter's dependency on company settings (e.g.
// the admin email a dispute alert is sent to) without wiring a full settings
// service into this integration test package.
type stubSettingsProvider struct{}

func (s *stubSettingsProvider) GetCompanyEmail(ctx context.Context) (string, error) {
	return "admin@leviosa.test", nil
}
func (s *stubSettingsProvider) GetCompanyLegalAddress(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyInstagram(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyLogo(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyTelephone(ctx context.Context) (string, error) {
	return "", nil
}
func (s *stubSettingsProvider) GetCompanyName(ctx context.Context) (string, error) {
	return "Leviosa Care", nil
}
func (s *stubSettingsProvider) InvalidateCache(key string) {}
func (s *stubSettingsProvider) InvalidateAllCache()        {}
