import { supabase } from "../config/config.js";

try {
  const sqlfile = Bun.file("./schema.sql");
  const sql = await sqlfile.text();

  console.log("Setting up the database schema...");

  // Split SQL by semicolon
  const statements = sql
    .split(";")
    .map(s => s.trim())
    .filter(s => s.length > 0);

  for (const stmt of statements) {
    console.log("Executing:", stmt.substring(0, 40), "...");
    const { error } = await supabase.rpc("execute_sql", { sql: stmt });

    if (error) throw error;
  }

  console.log("Database schema applied successfully!");
} catch (error) {
  console.error("Error setting up the database schema:", error);
}
