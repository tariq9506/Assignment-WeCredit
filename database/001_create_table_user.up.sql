


CREATE TABLE "user" (
  "id" SERIAL PRIMARY KEY,
  "phone_number" VARCHAR(15) UNIQUE,
  "otp" VARCHAR(4),
  "otp_valid_until" timestamp with time zone,
  "ip" INET,
  "location" TEXT,
  "phone_verified" BOOLEAN DEFAULT FALSE,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE "user_auth" (
  "id" SERIAL PRIMARY KEY,
  "jwt_token" TEXT,
  "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  "valid_until" timestamp with time zone,
  "user_id" BIGINT NOT NULL REFERENCES "user"("id"),
  "device_info" TEXT,
  "browser" TEXT,
  "ip" INET,
  "location" TEXT,
  "is_active" BOOLEAN DEFAULT TRUE
  );

CREATE TABLE countries (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    iso3 character(3),
    iso2 character(2),
    phonecode character varying(255),
    capital character varying(255),
    currency character varying(255),
    native character varying(255),
    emoji character varying(191),
    emojiu character varying(191),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    flag boolean DEFAULT true NOT NULL,
    wikidataid character varying(255)
);
INSERT INTO countries (id, name, iso2, phonecode)
VALUES
    (1, 'Japan', 'JP', '+81'),
    (2, 'India', 'IN', '+91'),
    (3, 'United State', 'US', '+1');
