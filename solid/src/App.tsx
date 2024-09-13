import { NavTopConfig1 } from "./NavTop";
import { Navigate } from "@solidjs/router";
import { ContextState, contextState } from "./context_state";
import { ApiEndpoint, useState } from "./state";
import { PageActivity } from "./page/activity";
import { PageMetric } from "./page/metric";
import { Router, Route } from "@solidjs/router";
import {
  createEffect,
  createResource,
  getOwner,
  JSX,
  onCleanup,
  runWithOwner,
} from "solid-js";
import {
  queryEndpointData,
  queryDatabaseInstanceInfo,
  queryDatabaseList,
  queryEndpointDataIfLive,
} from "./http";
import { Dynamic } from "solid-js/web";
import { DarkmodeSelector } from "./view/darkmode";
import { TimebarSection } from "./view/timebar_section";

export default function App(): JSX.Element {
  const { setState } = useState();
  createResource(() => queryDatabaseInstanceInfo(setState));
  const [databaseListIsReady] = createResource(() =>
    queryDatabaseList(setState),
  );
  const databaseIsReady = () => databaseListIsReady();

  return (
    <div class="max-w-screen-xl mx-auto">
      <ContextState>
        <Router>
          <Route path="/" component={() => <Navigate href="/activity" />} />
          <Route
            path="/activity"
            component={PageWrapper.bind(
              null,
              ApiEndpoint.activity,
              "pageActivity",
              PageActivity,
              databaseIsReady,
            )}
          />
          <Route
            path="/metric"
            component={PageWrapper.bind(
              null,
              ApiEndpoint.metric,
              "pageMetric",
              PageMetric,
              databaseIsReady,
            )}
          />
          <Route path={"**"} component={() => <h1>404. Page not found.</h1>} />
        </Router>
      </ContextState>
    </div>
  );
}

function PageWrapper(
  apiEndpoint: ApiEndpoint,
  testid: string,
  page: any,
  databaseIsReady: any,
) {
  const owner = getOwner();
  const { state, setState } = contextState();
  let timeout: any;
  let destroyed = false;

  function queryData() {
    runWithOwner(owner, () => {
      // runWithOwner needed because called by setTimeout
      createResource(databaseIsReady, () =>
        queryEndpointDataIfLive(apiEndpoint, state, setState),
      );
    });
    if (!destroyed) {
      timeout = setTimeout(queryData, state.interval_ms);
    }
  }

  createEffect(() => {
    // console.log("EFFECT interval", state.interval_ms);
    state.interval_ms;
    if (timeout) clearTimeout(timeout);
    queryData();
  });

  onCleanup(() => {
    destroyed = true;
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
  });

  return (
    <>
      <section data-testid={testid} class="flex flex-col mx-1 xs:mx-8">
        <NavTopConfig1 />
        <DatabaseHeader class="" />
        <Dynamic component={page} />
        <section class="sticky bottom-0 flex flex-col mt-3 z-20 backdrop-blur">
          <TimebarSection />
        </section>
      </section>
      <DarkmodeSelector class="mt-16 mb-4 self-start" />
    </>
  );
}

interface PropsDatabaseHeader {
  class?: string;
}

function DatabaseHeader(props: PropsDatabaseHeader) {
  const { state } = contextState();
  return (
    <section
      data-testid="db-header"
      class={`flex flex-col gap-y-1 ${props.class}`}
    >
      <h1 class="text-lg xs:text-2xl font-semibold">
        {state.database_instance.dbidentifier}
      </h1>
      <p class="text-neutral-500 flex flex-wrap gap-x-4 dark:text-neutral-400 text-xs sm:text-sm">
        <span>{state.database_instance.engine}</span>
        <span>{state.database_instance.engine_version}</span>
        <span>{state.database_instance.instance_class}</span>
      </p>
    </section>
  );
}
