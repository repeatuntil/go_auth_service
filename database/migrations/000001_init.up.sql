CREATE OR REPLACE FUNCTION get_future_date(days integer)
RETURNS timestamp with time zone
LANGUAGE 'plpgsql'
AS $$
	BEGIN
		RETURN NOW() + INTERVAL '1 day' * days;
	END;                
$$;

CREATE TABLE IF NOT EXISTS "refresh_token"
(
    id uuid PRIMARY KEY NOT NULL,
    userid uuid,
    refreshtoken bytea NOT NULL,
    ip varchar(15) COLLATE pg_catalog."default" NOT NULL,
    createdat timestamp with time zone NOT NULL DEFAULT now(),
    expiresat timestamp with time zone NOT NULL DEFAULT get_future_date(30),
    useragent varchar(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS inx_sessions_userid ON refresh_token(userid);
CREATE INDEX IF NOT EXISTS inx_sessions_expiresat ON refresh_token(expiresat);
