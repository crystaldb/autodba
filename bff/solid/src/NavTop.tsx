/** @import {JSX} from 'solid-js' */

// import { contextState } from "../context_state";

import { A } from "@solidjs/router";
import type { JSX } from "solid-js";

export function NavTop(props: {
  class?: string;
  children:
    | number
    | boolean
    | Node
    | JSX.ArrayElement
    | (string & {})
    | null
    | undefined;
}) {
  // const { state } = contextState();
  return (
    <>
      <nav
        class={`flex flex-col sm:flex-row items-center justify-between pe-3 border-b border-zinc-500 ${props.class}`}
      >
        <A href="/" class="flex items-center gap-2" end>
          <img src="/logo.svg" alt="logo" class="h-7" />
          <span class="text-2xl font-medium">AutoDBA</span>
        </A>
        <section class="flex flex-wrap gap-4 items-center">
          {props.children}
        </section>
      </nav>
    </>
  );
}

export function NavTopConfig1() {
  return (
    <NavTop class="mb-8">
      <A
        activeClass="activeTopNav"
        href="/activity"
        class="flex items-center justify-center h-16 px-4"
        end
      >
        Activity
      </A>
      <A
        activeClass="activeTopNav"
        href="/metrics"
        class="flex items-center justify-center h-16 px-4"
        end
      >
        Metrics
      </A>
      <A
        activeClass="activeTopNav"
        href="/prometheus"
        class="flex items-center justify-center h-16 px-4"
        end
      >
        Prometheus
      </A>
      {/*
      <div class="h-5 border-s w-1 border-neutral-200 dark:border-neutral-700"></div>
      <A
        activeClass="activeTopNav"
        href="/health"
        class="flex items-center justify-center h-16 px-4"
        end
      >
        Health
      </A>
      <A
        activeClass="activeTopNav"
        href="/explorer"
        class="flex items-center justify-center h-16 px-4"
        end
      >
        Explorer
      </A>
      */}
    </NavTop>
  );
}
