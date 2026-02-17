CREATE TABLE "_params" (
    "id"      TEXT PRIMARY KEY DEFAULT ('r'||substring(md5(random()::text) from 1 for 14)) NOT NULL,
    "value"   JSON DEFAULT NULL,
    "created" TEXT DEFAULT '' NOT NULL,
    "updated" TEXT DEFAULT '' NOT NULL
);

CREATE TABLE "collections" (
    "id"         TEXT PRIMARY KEY DEFAULT ('r'||substring(md5(random()::text) from 1 for 14)) NOT NULL,
    "system"     BOOLEAN DEFAULT FALSE NOT NULL,
    "type"       TEXT DEFAULT 'base' NOT NULL,
    "name"       TEXT UNIQUE NOT NULL,
    "fields"     JSON DEFAULT '[]' NOT NULL,
    "indexes"    JSON DEFAULT '[]' NOT NULL,
    "listRule"   TEXT DEFAULT NULL,
    "viewRule"   TEXT DEFAULT NULL,
    "createRule" TEXT DEFAULT NULL,
    "updateRule" TEXT DEFAULT NULL,
    "deleteRule" TEXT DEFAULT NULL,
    "options"    JSON DEFAULT '{}' NOT NULL,
    "created"    TEXT DEFAULT '' NOT NULL,
    "updated"    TEXT DEFAULT '' NOT NULL
);

CREATE INDEX IF NOT EXISTS "idx__collections_type" on "collections" ("type");
