/* =======================================================================
   HOW TO USE THIS SCHEMA
   =======================================================================
    This file contains the SQL commands to create the necessary tables for your project.
    WARNING: Running this script on an existing database may overwrite existing tables with the same names.
    Please ensure you have backups of your data before proceeding.
    Please replace "xxx" with your actual table names in the script below and remove all this comment to aviod error during sql execution.

   You can create these tables in THREE ways:

   -----------------------------------------------------------------------
   1. Using Supabase SQL Editor (RECOMMENDED)
   -----------------------------------------------------------------------
   - Open Supabase Dashboard → SQL Editor
   - Paste this entire file
   - Click RUN
   Done.

   -----------------------------------------------------------------------
   2. Using Supabase CLI (for migrations)
   -----------------------------------------------------------------------
   - npm install -g supabase
   - supabase login
   - supabase migration new init_schema
   - paste this file into /supabase/migrations/xxx_init_schema.sql
   - run: supabase db push

   -----------------------------------------------------------------------
   3. Using Bun RPC (Automated / CI / GitHub Actions)
   -----------------------------------------------------------------------
   FIRST create RPC function:
     CREATE OR REPLACE FUNCTION public.execute_sql(sql text)
     RETURNS void AS $$
     BEGIN
       EXECUTE sql;
     END;
     $$ LANGUAGE plpgsql SECURITY DEFINER;

     GRANT EXECUTE ON FUNCTION public.execute_sql(text) TO service_role;

   THEN run:
     bun --env-file=.env run setup-db.ts

   =======================================================================
   REPLACE "xxx" IN THIS FILE WITH YOUR ACTUAL TABLE NAMES:
   -----------------------------------------------------------------------
     API KEYS TABLE       =   api_keys
     REPORTS TABLE        =   reports
     DAILY CACHE TABLE    =   summary
   =======================================================================
*/


/* =======================================================================
   1. API KEYS TABLE  (replace xxx → your API keys table)
   ======================================================================= */

CREATE TABLE IF NOT EXISTS xxx (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  key_hash VARCHAR(64) NOT NULL UNIQUE,
  key_prefix VARCHAR(20) NOT NULL,
  is_active BOOLEAN DEFAULT true,
  rate_limit INTEGER DEFAULT 100000,
  usage_count INTEGER DEFAULT 0,
  last_used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_xxx_hash ON xxx(key_hash);
CREATE INDEX IF NOT EXISTS idx_xxx_active ON xxx(is_active);


/* =======================================================================
   2. REPORTS TABLE (Monitoring Reports)
   ======================================================================= */

CREATE TABLE IF NOT EXISTS xxx (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  service VARCHAR(255) NOT NULL,
  environment VARCHAR(100) DEFAULT 'production',
  total_checks INTEGER NOT NULL,
  uptime_count INTEGER NOT NULL,
  downtime_count INTEGER NOT NULL,
  degraded_count INTEGER NOT NULL,
  uptime_percent DECIMAL(5, 2) NOT NULL,
  average_latency_ms DECIMAL(10, 2) NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL,
  results JSONB NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_xxx_environment ON xxx(environment);
CREATE INDEX IF NOT EXISTS idx_xxx_timestamp ON xxx(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_xxx_created_at ON xxx(created_at DESC);

/* FIXED: JSON domain index (valid operator class) */
CREATE INDEX IF NOT EXISTS idx_xxx_results_domain
  ON xxx
  USING GIN ((results -> 'domain') jsonb_path_ops);


/* =======================================================================
   3. DAILY SUMMARY CACHE TABLE
   ======================================================================= */

CREATE TABLE IF NOT EXISTS xxx (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  domain VARCHAR(255) NOT NULL,
  date DATE NOT NULL,
  status VARCHAR(50) NOT NULL,
  title VARCHAR(255),
  description TEXT,
  time_down INTERVAL,
  total_downtime INTEGER DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(domain, date)
);

CREATE INDEX IF NOT EXISTS idx_xxx_domain_date ON xxx(domain, date);
CREATE INDEX IF NOT EXISTS idx_xxx_status ON xxx(status);


/* =======================================================================
   4. UPDATE TIMESTAMP TRIGGER
   ======================================================================= */

CREATE OR REPLACE FUNCTION update_xxx_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_xxx_updated_at_xxx
  BEFORE UPDATE ON xxx
  FOR EACH ROW
  EXECUTE FUNCTION update_xxx_updated_at();

CREATE OR REPLACE FUNCTION update_xxx_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_xxx_updated_at_xxx
  BEFORE UPDATE ON xxx
  FOR EACH ROW
  EXECUTE FUNCTION update_xxx_updated_at();

CREATE OR REPLACE FUNCTION update_xxx_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_xxx_updated_at_xxx
  BEFORE UPDATE ON xxx
  FOR EACH ROW
  EXECUTE FUNCTION update_xxx_updated_at();

/* =======================================================================
   5. ENABLE RLS
   ======================================================================= */

ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;
ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;
ALTER TABLE xxx ENABLE ROW LEVEL SECURITY;


/* =======================================================================
   6. RLS POLICIES (SERVICE ROLE FULL ACCESS)
   ======================================================================= */

CREATE POLICY "xxx full access - service_role"
  ON xxx
  FOR ALL
  USING (auth.role() = 'service_role');

CREATE POLICY "xxx full access - service_role"
  ON xxx
  FOR ALL
  USING (auth.role() = 'service_role');

CREATE POLICY "xxx full access - service_role"
  ON xxx
  FOR ALL
  USING (auth.role() = 'service_role');
