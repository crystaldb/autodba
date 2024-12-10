import { batch } from "solid-js";
import { fetchWithAuth } from "~/api";
import { contextState } from "~/context_state";

// BEGIN HACK CODE: (TODO replace this with SolidQuery)
// we are (temporarily) prioritizing time-to-completion over quality here while we work out the product spec for retries and exponential backoffs.
/** globalWithTemporaryHackTimeouts
 * CONTEXT: This code is temporary until we implement retry with exponential backoff. A global variable is used because, during development, each time this file is saved, Vite reloads the code, getting around the check to see if a timeout already exists. As a result, a developer can quickly have tons of requests being retried every 5 seconds, causing requests to continually be sending, which turns your laptop fan on. So, for now, we're polluting the global namespace since this code will be removed soon.
 */
const retryMs = 5000;
type GlobalWithTemporaryHackTimeouts = typeof globalThis & {
  timeout_queryDatabases: NodeJS.Timeout | null;
  timeout_queryInstances: NodeJS.Timeout | null;
};
const globalWithTemporaryHackTimeouts =
  globalThis as GlobalWithTemporaryHackTimeouts;
function retryQuery(
  blockExcessRetriesKey: "timeout_queryDatabases" | "timeout_queryInstances",
  fn: (arg0: boolean) => Promise<boolean>,
): boolean {
  if (globalWithTemporaryHackTimeouts[blockExcessRetriesKey]) return false;
  console.log(`Query: ${blockExcessRetriesKey}: will retry in ${retryMs}ms`);
  globalWithTemporaryHackTimeouts[blockExcessRetriesKey] = setTimeout(() => {
    globalWithTemporaryHackTimeouts[blockExcessRetriesKey] = null;
    fn(true);
  }, retryMs);
  return false;
}
// END HACK CODE

export async function queryInstances(retryIfNeeded: boolean): Promise<boolean> {
  const { setState } = contextState();
  const response = await fetchWithAuth("/api/v1/instance", { method: "GET" });

  if (!response.ok) {
    if (retryIfNeeded)
      return retryQuery("timeout_queryInstances", queryInstances);
    return false;
  }
  const json = await response.json();
  const instance_list = json?.list || [];
  if (!instance_list.length) {
    if (retryIfNeeded)
      return retryQuery("timeout_queryInstances", queryInstances);
    return false;
  }
  const instance_active = instance_list[0]
    ? JSON.parse(JSON.stringify(instance_list[0]))
    : null;
  batch(() => {
    setState("instance_active", instance_active);
    setState("instance_list", [
      ...instance_list,
      // {
      //   dbIdentifier: "0000000000111111111222222222233333333334444444444455555555555" + "::" + "amazon_rds" + "::" + "us-west-99",
      //   systemId: "0000000000111111111222222222233333333334444444444455555555555",
      //   systemType: "amazon_rds",
      //   systemScope: "us-west-99",
      // },
    ]);
  });
  if (retryIfNeeded) queryDatabases(retryIfNeeded);
  return true;
}

async function queryDatabases(retryIfNeeded: boolean): Promise<boolean> {
  const { state, setState } = contextState();
  if (!state.instance_active?.dbIdentifier) {
    if (retryIfNeeded)
      return retryQuery("timeout_queryDatabases", queryDatabases);
    return false;
  }
  const response = await fetchWithAuth(
    `/api/v1/instance/database?dbidentifier=${state.instance_active.dbIdentifier}`,
    { method: "GET" },
  );

  if (!response.ok) {
    if (retryIfNeeded)
      return retryQuery("timeout_queryDatabases", queryDatabases);
    return false;
  }
  const json = await response.json();

  setState("database_list", json || []);
  return true;
}
