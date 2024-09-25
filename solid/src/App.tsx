import { NavTopConfig1 } from "./NavTop";
import { Navigate } from "@solidjs/router";
import { ContextState, contextState } from "./context_state";
import { PageActivity } from "./page/activity";
import { PageMetric } from "./page/metric";
import { Router, Route } from "@solidjs/router";
import { batch, createSignal, For, JSX, Show } from "solid-js";
import { queryInstances, queryDatabases } from "./http";
import { Dynamic } from "solid-js/web";
import { DarkmodeSelector } from "./view/darkmode";
import { TimebarSection } from "./view/timebar_section";
import {
  cssSelectorGeneral,
  cssSelectorGeneralBase,
  cssSelectorGeneralHover,
} from "./view/cube_activity";
import { VsTriangleDown } from "solid-icons/vs";
import { Instance } from "./state";

export default function App(): JSX.Element {
  return (
    <ContextState>
      <Router root={Layout}>
        <Route path="/" component={() => <Navigate href="/activity" />} />
        <Route
          path="/activity"
          component={() => PageWrapper("pageActivity", PageActivity)}
        />
        <Route
          path="/metric"
          component={() => PageWrapper("pageMetric", PageMetric)}
        />
        <Route path={"**"} component={() => <h1>404. Page not found.</h1>} />
      </Router>
    </ContextState>
  );
}

function Layout(props: {
  children?:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}): JSX.Element {
  queryInstances(true);
  queryDatabases(true);

  return <div class="max-w-screen-xl mx-auto">{props.children}</div>;
}

function PageWrapper(testid: string, page: () => JSX.Element) {
  const { state } = contextState();
  return (
    <>
      <section data-testid={testid} class="flex flex-col mx-1 xs:mx-8">
        <NavTopConfig1 />
        <DatabaseHeader class="mb-8" />
        <Show when={state.database_list.length}>
          <Dynamic component={page} />
          <section class="sticky bottom-0 flex flex-col mt-3 pt-1 z-20 backdrop-blur">
            <TimebarSection />
          </section>
        </Show>
      </section>
      <DarkmodeSelector class="mt-16 mb-4 self-start" />
    </>
  );
}

interface PropsDatabaseHeader {
  class?: string;
}
function DatabaseHeader(props: PropsDatabaseHeader) {
  const { state, setState } = contextState();
  const [selectInstanceOpen, setSelectInstanceOpen] =
    createSignal<boolean>(false);

  return (
    <section
      data-testid="db-header"
      class={`flex flex-col gap-y-1 ${props.class}`}
    >
      <Show
        when={state.instance_list.length > 1}
        fallback={
          <h1 class="text-lg xs:text-2xl font-semibold">
            {state.instance_active?.dbIdentifier || ""}
          </h1>
        }
      >
        <section class="relative self-start flex flex-col gap-y-1">
          <h1
            class={`text-lg xs:text-2xl font-semibold py-4 px-6 rounded-md flex items-center gap-4 cursor-pointer ${cssSelectorGeneral}`}
            onClick={() => setSelectInstanceOpen((prev) => !prev)}
          >
            <span>{state.instance_active?.dbIdentifier || ""}</span>
            <span>
              <VsTriangleDown size={24} />
            </span>
          </h1>
          <Show when={selectInstanceOpen()}>
            <ul
              class={`absolute top-16 mt-1 z-10 font-medium rounded-md divide-y divide-zinc-200 dark:divide-zinc-600 ${cssSelectorGeneralBase}`}
            >
              <For each={state.instance_list}>
                {(instance: Instance, index: () => number) => (
                  <li
                    class={`px-6 py-4 cursor-pointer ${cssSelectorGeneralHover}`}
                    classList={{
                      "text-fuchsia-500":
                        instance.dbIdentifier ===
                        state.instance_active?.dbIdentifier,
                    }}
                    onClick={() =>
                      batch(() => {
                        setState(
                          "instance_active",
                          state.instance_list[index()],
                        );
                        setSelectInstanceOpen(false);
                      })
                    }
                  >
                    <p>{instance.dbIdentifier}</p>
                    <p class="text-neutral-500 flex flex-wrap gap-x-4 dark:text-neutral-400 text-xs sm:text-sm">
                      {/*
                      <span>{instance.systemId}</span>
                      */}
                      <span>{instance.systemType}</span>
                      <span>{instance.systemScope}</span>
                    </p>
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </section>
      </Show>
      <p class="text-neutral-500 flex flex-wrap gap-x-4 dark:text-neutral-400 text-xs sm:text-sm">
        <span>{state.instance_active?.systemType || ""}</span>
        <span>{state.instance_active?.systemScope || ""}</span>
        <span>id:{state.instance_active?.systemId || ""}</span>
      </p>
    </section>
  );
}
