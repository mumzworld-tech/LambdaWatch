// LambdaWatch E2E Test Function
// TEST ONLY - can be safely deleted
// Generates various log patterns to verify the extension works correctly

export const handler = async (event) => {
  const testType = event.testType || "basic";

  console.log(JSON.stringify({ level: "info", message: "Test function invoked", testType }));

  switch (testType) {
    case "basic":
      return basicLogs();
    case "volume":
      return volumeLogs(event.count || 50);
    case "large":
      return largeLogs(event.sizeKB || 5);
    case "json":
      return jsonLogs();
    case "error":
      return errorLogs();
    case "slow":
      return slowLogs(event.durationMs || 3000);
    default:
      return basicLogs();
  }
};

function basicLogs() {
  console.log(JSON.stringify({ level: "info", message: "Starting processing" }));
  console.log(JSON.stringify({ level: "debug", message: "Loading configuration", config: { timeout: 30 } }));
  console.log(JSON.stringify({ level: "info", message: "Processing item", itemId: "item-001" }));
  console.log(JSON.stringify({ level: "info", message: "Processing item", itemId: "item-002" }));
  console.log(JSON.stringify({ level: "info", message: "Processing complete", itemsProcessed: 2 }));
  return { statusCode: 200, body: "basic: 5 logs generated" };
}

function volumeLogs(count) {
  for (let i = 0; i < count; i++) {
    console.log(JSON.stringify({ level: "info", message: `Log entry ${i + 1}/${count}`, index: i }));
  }
  return { statusCode: 200, body: `volume: ${count} logs generated` };
}

function largeLogs(sizeKB) {
  const padding = "x".repeat(sizeKB * 1024);
  console.log(JSON.stringify({ level: "info", message: "Large log entry", data: padding }));
  return { statusCode: 200, body: `large: 1 log of ~${sizeKB}KB generated` };
}

function jsonLogs() {
  console.log(JSON.stringify({
    level: "info",
    message: "User action",
    user: { id: "u-123", role: "admin" },
    action: "login",
    metadata: { ip: "10.0.0.1", userAgent: "test" },
  }));
  console.log(JSON.stringify({
    level: "warn",
    message: "Rate limit approaching",
    current: 95,
    limit: 100,
  }));
  console.log(JSON.stringify({
    level: "error",
    message: "Validation failed",
    errors: [{ field: "email", reason: "invalid format" }],
  }));
  return { statusCode: 200, body: "json: 3 structured logs generated" };
}

function errorLogs() {
  console.log(JSON.stringify({ level: "info", message: "Starting error test" }));
  try {
    throw new Error("Simulated error for testing");
  } catch (e) {
    console.error(JSON.stringify({ level: "error", message: e.message, stack: e.stack }));
  }
  console.log(JSON.stringify({ level: "info", message: "Error test complete" }));
  return { statusCode: 200, body: "error: 3 logs generated (including error)" };
}

async function slowLogs(durationMs) {
  const interval = durationMs / 5;
  for (let i = 0; i < 5; i++) {
    console.log(JSON.stringify({ level: "info", message: `Slow log ${i + 1}/5`, elapsed: i * interval }));
    await new Promise((r) => setTimeout(r, interval));
  }
  return { statusCode: 200, body: `slow: 5 logs over ${durationMs}ms` };
}
