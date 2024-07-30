import { PageHome } from "./page/home";
import NavTop from "./NavTop";
import { A } from "@solidjs/router";
import { ContextState } from "./context_state";
import { useState } from "./state";
import { PageAll } from "./page/all";
import { Router, Route } from "@solidjs/router";
import { createResource, JSX } from "solid-js";
import { getData } from "./http";
import { HiSolidChevronLeft, HiSolidChevronRight } from "solid-icons/hi";

export default function App(): JSX.Element {
  const { setState } = useState();
  createResource(getData.bind(null, setState));

  return (
    <div class="max-w-screen-sm mx-auto">
      <ContextState>
        <Router>
          <Route path="/" component={PageHomeWrapper} />
          <Route path={`/data/all`} component={PageAllWrapper} />
          {/*<Route path={"**"} component=<h1>404. Page not found.</h1> />*/}
        </Router>
      </ContextState>
    </div>
  );
}

function PageHomeWrapper() {
  const testid = "pageHome";
  return (
    <section data-testid={testid} class="flex flex-col">
      <NavTop>
        <A
          activeClass="active"
          role="link"
          href="/data/all"
          class="flex items-center justify-center w-10 h-full"
          end
        >
          <HiSolidChevronRight size={24} />
        </A>
      </NavTop>

      <h1 class="text-center text-4xl font-semibold mb-8">
        Crystal Observability
      </h1>

      <PageHome />
    </section>
  );
}

function PageAllWrapper() {
  const testid = "pageAll";
  return (
    <section data-testid={testid} class="flex flex-col">
      <NavTop>
        <A
          activeClass="active"
          href="/"
          class="flex items-center justify-center w-10 h-full"
          end
        >
          <HiSolidChevronLeft size={24} />
        </A>
      </NavTop>

      <h1 class="text-center text-4xl font-semibold mt-2 mb-6 flex flex-col gap-y-2">
        All Charts
      </h1>

      <PageAll />
    </section>
  );
}
