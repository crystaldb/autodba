/** @import {JSX} from 'solid-js' */

// import { contextState } from "../context_state";

import { A } from "@solidjs/router";
import { JSX } from "solid-js";

function NavTop(props: {
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
  // <img src="/logo.webp" alt="logo" class="bg-white w-[652px] h-[135px]" />
  // divided by 3 = 217px x 45px
  // divided by 4 = 163px x 34px
  return (
    <>
      <nav class="flex items-center justify-between h-16 px-2">
        <A href="/" class="block dark:hidden">
          <img
            src="/logo-dark-text.svg"
            alt="logo"
            class="w-[217px] h-[45px]"
          />
        </A>
        <A href="/" class="hidden dark:block">
          <img
            src="/logo-light-text.svg"
            alt="logo"
            class="w-[217px] h-[45px]"
          />
        </A>
        {/*
        <section class="flex gap-x-4 items-center">{props.children}</section>
        */}
      </nav>
    </>
  );
}

export default NavTop;
