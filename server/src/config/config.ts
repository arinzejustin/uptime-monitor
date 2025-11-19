import { createClient } from "@supabase/supabase-js";

export const SUPABASE_URL = process.env.SUPABASE_URL!;
export const SUPABASE_KEY = process.env.SUPABASE_KEY!;
export const REPORTS_TABLE = process.env.REPORTS_TABLE || 'xxx';
export const API_KEY_TABLE = process.env.API_KEY_TABLE || 'xxx';
export const SUMMARY_TABLE = process.env.SUMMARY_TABLE || 'xxx';

export const supabase = createClient(SUPABASE_URL, SUPABASE_KEY);

