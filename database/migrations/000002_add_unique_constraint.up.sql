ALTER TABLE refresh_token
ADD CONSTRAINT refresh_session_unique UNIQUE (userid, useragent);
