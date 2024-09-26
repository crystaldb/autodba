import { NavTopConfig1 } from "./NavTop";
import { Navigate } from "@solidjs/router";
import { ContextState, contextState } from "./context_state";
import { PageActivity } from "./page/activity";
import { PageMetric } from "./page/metric";
import { Router, Route } from "@solidjs/router";
import { JSX, Show } from "solid-js";
import { queryInstances, queryDatabases } from "./http";
import { Dynamic } from "solid-js/web";
import { DarkmodeSelector } from "./view/darkmode";
import { TimebarSection } from "./view/timebar_section";

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
        <DatabaseHeader class="" />
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
