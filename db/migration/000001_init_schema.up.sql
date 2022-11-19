CREATE TABLE "users" (
  "username" varchar UNIQUE PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT('0001-01-01 00:00:00Z'),  
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "articles" (
  "id" bigserial PRIMARY KEY,
  "author" varchar NOT NULL,
  "headline" varchar NOT NULL,
  "content" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "edited_at" timestamptz
);

ALTER TABLE "articles" ADD FOREIGN KEY ("author") REFERENCES "users" ("username");
CREATE INDEX ON "articles" ("author");
