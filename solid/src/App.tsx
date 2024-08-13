import NavTop from "./NavTop";
import { A, Navigate } from "@solidjs/router";
import { ContextState, contextState } from "./context_state";
import { useState } from "./state";
import { PageHealth } from "./page/health";
import { PageActivity } from "./page/activity";
import { PageMetric } from "./page/metric";
import { PageExplorer } from "./page/explorer";
import { Router, Route } from "@solidjs/router";
import {
  createResource,
  createSignal,
  For,
  JSX,
  onCleanup,
  onMount,
  Show,
} from "solid-js";
import { getData, getDatabaseInfo } from "./http";
import { Dynamic } from "solid-js/web";
import { DarkmodeSelector } from "./view/darkmode";
import {
  HiOutlineChevronDown,
  HiOutlineChevronLeft,
  HiSolidArrowDown,
} from "solid-icons/hi";

export default function App(): JSX.Element {
  const { setState } = useState();
  const getDataFn = getDatabaseInfo.bind(null, setState);
  createResource(getDataFn);

  return (
    <div class="max-w-screen-xl mx-auto">
      <ContextState>
        <Router>
          <Route path="/" component={() => <Navigate href="/health" />} />
          <Route
            path="/health"
            component={PageWrapper.bind(
              null,
              "health",
              "pageHealth",
              PageHealth
            )}
          />
          <Route
            path="/activity"
            component={PageWrapper.bind(
              null,
              "activity",
              "pageActivity",
              PageActivity
            )}
          />
          <Route
            path="/metric"
            component={PageWrapper.bind(
              null,
              "metric",
              "pageMetric",
              PageMetric
            )}
          />
          <Route
            path="/explorer"
            component={PageWrapper.bind(
              null,
              "explorer",
              "pageExplorer",
              PageExplorer
            )}
          />
          <Route path={"**"} component={() => <h1>404. Page not found.</h1>} />
        </Router>
      </ContextState>
    </div>
  );
}

function PageWrapper(apiEndpoint: string, testid: string, page: any) {
  const { state, setState } = contextState();
  const getDataFn = getData.bind(null, apiEndpoint, setState);
  let timeout: any;
  let destroyed = false;

  onMount(() => {
    function queryData() {
      createResource(getDataFn);
      if (!destroyed) {
        timeout = setTimeout(queryData, state.interval_ms);
      }
    }
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
    <section data-testid={testid} class="flex flex-col mx-1 xs:mx-8">
      <NavTopConfig1 />
      <DatabaseHeader class="mb-4" />
      <IntervalSelector class="self-start my-4" />
      <Dynamic component={page} />
      <DarkmodeSelector class="mt-16 mb-4" />
    </section>
  );
}

function DatabaseHeader(props: { class?: any }) {
  const { state } = contextState();
  return (
    <section class={`flex flex-col gap-y-1 ${props.class}`}>
      <h1 class="text-2xl font-semibold">{state.database.name}</h1>
      <p class="text-neutral-500 flex flex-wrap gap-x-4 text-neutral-600 dark:text-neutral-400 text-sm">
        <span>{state.database.engine}</span>
        <span>{state.database.version}</span>
        <span>{state.database.size}</span>
        <span>{state.database.kind}</span>
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
        <label>Interval:</label>
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
        href="/health"
        class="flex items-center justify-center h-full"
        end
      >
        Health
      </A>
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
      <A
        activeClass="active"
        href="/explorer"
        class="flex items-center justify-center h-full"
        end
      >
        Explorer
      </A>
    </NavTop>
  );
}
