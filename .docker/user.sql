CREATE TABLE "user" (
  "id" uuid PRIMARY KEY,
  "email" character varying(254) UNIQUE NOT NULL,
  "password" character(60) NOT NULL
);