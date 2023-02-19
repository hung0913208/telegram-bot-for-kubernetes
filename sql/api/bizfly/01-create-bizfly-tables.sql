

DROP TABLE IF EXISTS "public"."tbl_bizfly_account";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_account" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "email" text,
    "password" text,
    "project_id" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_account_idx_email
ON public.tbl_bizfly_account USING BTREE
(
    "email" ASC
);


DROP TABLE IF EXISTS "public"."tbl_bizfly_cluster";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_cluster" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "account" text,
    "name" text,
    "status" text,
    "balance" int8,
    "locked" bool,
    "tags" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_cluster_idx_account_id
ON public.tbl_bizfly_cluster USING BTREE
(
    "account" ASC
);

CREATE INDEX tbl_bizfly_cluster_idx_name
ON public.tbl_bizfly_cluster USING BTREE
(
    "name" ASC
);

CREATE INDEX tbl_bizfly_cluster_idx_locked
ON public.tbl_bizfly_cluster USING BTREE
(
    "locked" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_cluster_stat";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_cluster_stat" (
    "cluster" text NOT NULL,
    "account" text,
    "core" int8,
    "memory" int8,
    PRIMARY KEY ("cluster")
);

CREATE INDEX tbl_bizfly_cluster_stat_idx_account_id
ON public.tbl_bizfly_cluster_stat USING BTREE
(
	"account" ASC
);


DROP TABLE IF EXISTS "public"."tbl_bizfly_firewall";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_firewall" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "account" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_firewall_idx_account_id
ON public.tbl_bizfly_firewall USING BTREE
(
    "account" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_firewall_bound";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_firewall_bound" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "account" text,
    "firewall" text,
    "type" int8,
    "c_id_r" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_firewall_bound_idx_account_id
ON public.tbl_bizfly_firewall_bound USING BTREE
(
    "account" ASC
);

CREATE INDEX tbl_bizfly_firewall_bound_idx_firewall_id
ON public.tbl_bizfly_firewall_bound USING BTREE
(
    "firewall" ASC
);

CREATE INDEX tbl_bizfly_firewall_bound_idx_bound_type
ON public.tbl_bizfly_firewall_bound USING BTREE
(
    "type" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_pool";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_pool" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "name" text,
    "account" text,
    "cluster" text,
    "zone" text,
    "status" text,
    "autoscale" text,
    "enable_autoscaling" bool,
    "required_size" int8,
    "limit_size" int8,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_pool_idx_name
ON public.tbl_bizfly_pool USING BTREE
(
	"name" ASC
);

CREATE INDEX tbl_bizfly_pool_idx_account_id
ON public.tbl_bizfly_pool USING BTREE
(
	"account" ASC
);

CREATE INDEX tbl_bizfly_pool_idx_cluster_id
ON public.tbl_bizfly_pool USING BTREE
(
	"cluster" ASC
);

CREATE INDEX tbl_bizfly_pool_idx_zone
ON public.tbl_bizfly_pool USING BTREE
(
	"zone" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_pool_node";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_pool_node" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "name" text,
    "account" text,
    "pool" text,
    "cluster" text,
    "server" text,
    "status" text,
    "reason" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_pool_node_idx_account_id
ON public.tbl_bizfly_pool_node USING BTREE
(
	"account" ASC
);

CREATE INDEX tbl_bizfly_pool_node_idx_cluster_id
ON public.tbl_bizfly_pool_node USING BTREE
(
	"cluster" ASC
);

CREATE INDEX tbl_bizfly_pool_node_idx_server_id
ON public.tbl_bizfly_pool_node USING BTREE
(
	"server" ASC
);

CREATE INDEX tbl_bizfly_pool_node_idx_status
ON public.tbl_bizfly_pool_node USING BTREE
(
	"status" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_server";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_server" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "account" text,
    "status" text,
    "cluster" text,
    "balance" int8,
    "locked" bool,
    "zone" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_server_idx_account_id
ON public.tbl_bizfly_server USING BTREE
(
	"account" ASC
);

CREATE INDEX tbl_bizfly_server_idx_status
ON public.tbl_bizfly_server USING BTREE
(
	"status" ASC
);

CREATE INDEX tbl_bizfly_server_idx_cluster_id
ON public.tbl_bizfly_server USING BTREE
(
	"cluster" ASC
);

CREATE INDEX tbl_bizfly_server_idx_locked
ON public.tbl_bizfly_server USING BTREE
(
	"locked" ASC
);

CREATE INDEX tbl_bizfly_server_idx_zone
ON public.tbl_bizfly_server USING BTREE
(
	"zone" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_volume";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_volume" (
    "uuid" text NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "account" text,
    "type" text,
    "description" text,
    "status" text,
    "zone" text,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_bizfly_volume_idx_account_id
ON public.tbl_bizfly_volume USING BTREE
(
    "account" ASC
);


DROP TABLE IF EXISTS "public"."tbl_bizfly_volume_cluster";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_volume_cluster" (
    "volume" text NOT NULL,
    "account" text NOT NULL,
    "pod" text NOT NULL,
    "cluster" text NOT NULL,
    "size" int8,
    PRIMARY KEY ("pod","cluster")
);

CREATE INDEX tbl_bizfly_volume_cluster_idx_cluster
ON public.tbl_bizfly_volume_cluster USING BTREE
(
    "cluster" ASC
);

CREATE INDEX tbl_bizfly_volume_cluster_idx_pod
ON public.tbl_bizfly_volume_cluster USING BTREE
(
    "pod" ASC
);

CREATE INDEX tbl_bizfly_volume_cluster_idx_volume_id
ON public.tbl_bizfly_volume_cluster USING BTREE
(
    "volume" ASC
);

CREATE INDEX tbl_bizfly_volume_cluster_idx_account_id
ON public.tbl_bizfly_volume_cluster USING BTREE
(
    "account" ASC
);

DROP TABLE IF EXISTS "public"."tbl_bizfly_volume_server";
-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_bizfly_volume_server" (
    "volume" text NOT NULL,
    "account" text,
    "server" text,
    PRIMARY KEY ("volume")
);

CREATE INDEX tbl_bizfly_volume_server_idx_account_id
ON public.tbl_bizfly_volume_server USING BTREE
(
    "account" ASC
);

CREATE INDEX tbl_bizfly_volume_server_idx_server_id
ON public.tbl_bizfly_volume_server USING BTREE
(
    "server" ASC
);

