ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_totp_secret;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_totp_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_sms_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_email_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS mfa_backup_codes;
