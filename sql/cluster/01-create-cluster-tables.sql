DROP TABLE IF EXISTS "public"."tbl_cluster_cluster";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_cluster" (
    "name"           varchar(30) NOT NULL,
    "created_at"     timestamptz,
    "updated_at"     timestamptz,
    "provider"       varchar(20) NOT NULL,
    "metadata"       text,
    "kubeconfig"     text,
    "expire"         int16,
    "infrastructure" text,
    PRIMARY KEY ("name")
);


DROP TABLE IF EXISTS "public"."tbl_clustter_alias";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_alias" (
    "alias"   varchar(30) NOT NULL,
    "cluster" varchar(30) NOT NULL,
    PRIMARY KEY ("alias")
);

CREATE INDEX tbl_cluster_alias_idx_cluster
ON public.tbl_cluster_alias USING BTREE
(
    "cluster" ASC
);


DROP TABLE IF EXISTS "public"."tbl_clustter_deployment";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_deployment" (
    "name"    varchar(30) NOT NULL,
    "cluster" varchar(30) NOT NULL,
    "kind"    varchar(30) NOT NULL,
    PRIMARY KEY ("name", "cluster")
);


CREATE INDEX tbl_cluster_deployment_idx_cluster
ON public.tbl_cluster_deployment USING BTREE
(
    "cluster" ASC
);

CREATE INDEX tbl_cluster_deployment_idx_kind
ON public.tbl_cluster_deployment USING BTREE
(
    "kind" ASC
);


DROP TABLE IF EXISTS "public"."tbl_clustter_pod";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_pod" (
    "uuid"           varchar(30) NOT NULL,
    "created_at"     timestamptz,
    "updated_at"     timestamptz,
    "deployment"     varchar(30) NOT NULL,
    "version"        varchar(50) NOT NULL,
    "status"         int8,
    "cpu_limit"      int32,
    "cpu_request"    int32,
    "memory_limit"   int32,
    "memory_request" int32,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_cluster_pod_idx_deployment
ON public.tbl_cluster_pod USING BTREE
(
    "deployment" ASC
);

CREATE INDEX tbl_cluster_pod_idx_status
ON public.tbl_cluster_pod USING BTREE
(
    "status" ASC
);

CREATE INDEX tbl_cluster_deployment_idx_cpu_request
ON public.tbl_cluster_deployment USING BTREE
(
    "cpu_request" ASC
);

CREATE INDEX tbl_cluster_deployment_idx_cpu_limit
ON public.tbl_cluster_deployment USING BTREE
(
    "cpu_limit" ASC
);

CREATE INDEX tbl_cluster_deployment_idx_memory_request
ON public.tbl_cluster_deployment USING BTREE
(
    "memory_request" ASC
);

CREATE INDEX tbl_cluster_deployment_idx_memory_limit
ON public.tbl_cluster_deployment USING BTREE
(
    "memory_limit" ASC
);


DROP TABLE IF EXISTS "public"."tbl_clustter_volume";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_volume" (
    "uuid"       varchar(30) NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "name"       varchar(30) NOT NULL,
    "cluster"    varchar(30) NOT NULL,
    "pod"  	 varchar(30) NOT NULL,
    "usage"      int64,
    "capacity"   int64,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_cluster_volume_pod
ON public.tbl_cluster_volume USING BTREE
(
    "pod" ASC
);

CREATE INDEX tbl_cluster_volume_usage
ON public.tbl_cluster_volume USING BTREE
(
    "usage" ASC
);


DROP TABLE IF EXISTS "public"."tbl_clustter_volume_snapshot";

-- This script only contains the table creation statements and does not fully represent the table in the database. It's still missing: indices, triggers. Do not use it as a backup.

-- Table Definition
CREATE TABLE "public"."tbl_cluster_volume_snapshot" (
    "uuid"       varchar(30) NOT NULL,
    "created_at" timestamptz,
    "updated_at" timestamptz,
    "cluster"    varchar(30) NOT NULL,
    "volume"     varchar(30) NOT NULL,
    "version"    int64,
    PRIMARY KEY ("uuid")
);

CREATE INDEX tbl_cluster_volume_snapshot_volume
ON public.tbl_cluster_volume_snapshot USING BTREE
(
    "volume" ASC
);

CREATE INDEX tbl_cluster_volume_snapshot_version
ON public.tbl_cluster_volume_snapshot USING BTREE
(
    "version" ASC
);
