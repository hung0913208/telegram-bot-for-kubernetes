
DROP TABLE IF EXISTS "public"."tbl_platform_backup";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_platform_backup" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "namespace" text,
    "volume" text,
    "image" text,
    "state" int8,
    PRIMARY KEY ("uuid")
);
