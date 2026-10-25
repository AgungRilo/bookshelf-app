-- +migrate Up
-- user default untuk login: admin / admin123 (password disimpan dalam bentuk bcrypt)
INSERT INTO users (username, password, created_by)
VALUES ('admin', '$2a$10$/sZjcxhIezkU5h/Cgxp3V.IY6kJm.u.BjutaUQvn6wErIPFrm2Q1G', 'system')
ON CONFLICT (username) DO NOTHING;

-- +migrate Down
DELETE FROM users WHERE username = 'admin';
