import { NavTop } from "./NavTop";
import { A, Navigate } from "@solidjs/router";
import { ContextState, contextState } from "./context_state";
import { useState } from "./state";
import { PageHealth } from "./page/health";
import { PageActivity } from "./page/activity";
import { PageMetric } from "./page/metric";
import { PageExplorer } from "./page/explorer";
import { Router, Route } from "@solidjs/router";
import {
  createEffect,
  createResource,
  For,
  getOwner,
  JSX,
  onCleanup,
  runWithOwner,
} from "solid-js";
import {
  queryEndpointData,
  queryDatabaseInstanceInfo,
  queryDatabaseList,
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
            path="/health"
            component={PageWrapper.bind(
              null,
              "health",
              "pageHealth",
              PageHealth,
              databaseIsReady,
            )}
          />
          <Route
            path="/activity"
            component={PageWrapper.bind(
              null,
              "activity",
              "pageActivity",
              PageActivity,
              databaseIsReady,
            )}
          />
          <Route
            path="/metric"
            component={PageWrapper.bind(
              null,
              "metric",
              "pageMetric",
              PageMetric,
              databaseIsReady,
            )}
          />
          <Route
            path="/explorer"
            component={PageWrapper.bind(
              null,
              "explorer",
              "pageExplorer",
              PageExplorer,
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
  apiEndpoint: string,
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
      createResource(databaseIsReady, () =>
        queryEndpointData(apiEndpoint, state, setState),
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
        <section class="flex flex-wrap justify-between gap-4 mb-8">
          <DatabaseHeader class="" />
          <IntervalSelector class="self-start" />
        </section>
        <Dynamic component={page} />
        <section class="sticky bottom-0 flex flex-col mt-3 z-20 backdrop-blur">
          <TimebarSection class="w-full xs:w-10/12" />
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

function IntervalSelector(props: { class?: string }) {
  const { state, setState } = contextState();
  // const [openGet, openSet] = createSignal(false);

  // <Show
  //   when={openGet()}
  //   fallback={<HiOutlineChevronLeft size="1rem" class="inline ms-1" />}
  // >
  //   <HiOutlineChevronDown size="1rem" class="inline ms-1" />
  // </Show>
  // <button
  //   onClick={() => openSet((v) => !v)}
  //   class={`inline ${props.class}`}
  // >
  //   Interval: {state.interval_ms} ms
  // </button>
  return (
    <>
      <div class={`flex items-center gap-x-3 text-sm ${props.class}`}>
        <label>Interval</label>
        <select
          onChange={(e) => setState({ interval_ms: +e.currentTarget.value })}
          class="bg-transparent rounded border border-neutral-200 dark:border-neutral-700 text-fuchsia-500 ps-2 pe-8 py-1.5 hover:border-gray-400 focus:outline-none"
        >
          <option
            value="5000"
            class="appearance-none bg-neutral-100 dark:bg-neutral-800"
          >
            Auto
          </option>
          <For
            each={[
              [1000, "1s"],
              [5000, "5s"],
              [10000, "10s"],
              [20000, "20s"],
              [30000, "30s"],
              [60000, "60s"],
              [300000, "5m"],
              [900000, "15m"],
              [1800000, "30m"],
              [3600000, "60m"],
            ]}
          >
            {([value, label]) => (
              <option
                value={value}
                {...(value === state.interval_ms && { selected: true })}
                class="appearance-none bg-neutral-100 dark:bg-neutral-800"
              >
                {label}
              </option>
            )}
          </For>
        </select>
      </div>
    </>
  );
}

function NavTopConfig1() {
  return (
    <NavTop class="mb-8">
      <A
        activeClass="active"
        href="/activity"
        class="flex items-center justify-center h-full"
        end
      >
        Activity
      </A>
      <A
        activeClass="active"
        href="/metric"
        class="flex items-center justify-center h-full"
        end
      >
        Metrics
      </A>
      {/*
      <div class="h-5 border-s w-1 border-neutral-200 dark:border-neutral-700"></div>
      <A
        activeClass="active"
        href="/health"
        class="flex items-center justify-center h-full"
        end
      >
        Health
      </A>
      <A
        activeClass="active"
        href="/explorer"
        class="flex items-center justify-center h-full"
        end
      >
        Explorer
      </A>
      */}
    </NavTop>
  );
}
