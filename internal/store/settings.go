package store

import "context"

// ProxyTemplate returns the encrypted global template. A nil value means
// proxy assignment is disabled.
func (s *Store) ProxyTemplate(ctx context.Context) ([]byte, error) {
	var encrypted []byte
	err := s.pool.QueryRow(ctx, `
		SELECT proxy_template_enc
		FROM admin_settings
		WHERE singleton = true`).Scan(&encrypted)
	return encrypted, wrapErr(err)
}

func (s *Store) SetProxyTemplate(ctx context.Context, encrypted []byte) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE admin_settings
		SET proxy_template_enc = $1, updated_at = now()
		WHERE singleton = true`, encrypted)
	return err
}

func (s *Store) ClearProxyTemplate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE admin_settings
		SET proxy_template_enc = NULL, updated_at = now()
		WHERE singleton = true`)
	return err
}
