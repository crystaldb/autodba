import type { Anything } from "~/state";

export function DebugJson(props: {
  // biome-ignore lint/suspicious/noExplicitAny: the goal is to show anything
  json: /*eslint-disable */ any /*eslint-enable */;
}) {
  return (
    <pre class="break-words max-w-96 overflow-auto text-xs">
      {JSON.stringify(props.json, null, 2)}
    </pre>
  );
}

export function ViewTitle(props: { title: Anything }) {
  return (
    <span class="font-medium dark:font-base sm:font-base sm:text-sm/6 leading-6">
      {props.title}
    </span>
  );
}
