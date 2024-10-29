import { contextState } from "../context_state";
import { batch, createSignal, For, Show } from "solid-js";
import {
  cssSelectorGeneral,
  cssSelectorGeneralBase,
  cssSelectorGeneralHover,
} from "../view/cube_activity";
import { VsTriangleDown } from "solid-icons/vs";
import { Instance } from "../state";

interface PropsInstanceHeader {
  class?: string;
}
export function InstanceHeader(props: PropsInstanceHeader) {
  const { state, setState } = contextState();
  const [selectInstanceOpen, setSelectInstanceOpen] =
    createSignal<boolean>(false);

  const cssMaxWidth =
    "max-w-[calc(100vw-2rem)] xs:max-w-[calc(min(100vw-5rem))]";

  return (
    <section
      data-testid="db-header"
      class={`flex flex-col gap-y-1 ${props.class}`}
    >
      <Show when={state.instance_list.length > 1} fallback=<H1Header />>
        <section class="relative self-start flex flex-col gap-y-1 max-w-[calc(100vw-2rem)] xs:max-w-[calc(min(100vw-5rem))]">
          <div
            class={` py-4 px-6 rounded-md flex items-center gap-4 cursor-pointer ${cssSelectorGeneral}`}
            onClick={() => setSelectInstanceOpen((prev) => !prev)}
          >
            <H1Header />
            <VsTriangleDown size={14} />
          </div>
          <Show when={selectInstanceOpen()}>
            <ul
              class={`absolute top-16 mt-1 z-10 ${cssMaxWidth} rounded-md divide-y divide-zinc-200 dark:divide-zinc-600 ${cssSelectorGeneralBase}`}
            >
              <For each={state.instance_list}>
                {(instance: Instance, index: () => number) => (
                  <li
                    class={`px-6 py-4 cursor-pointer overflow-auto ${cssSelectorGeneralHover}`}
                    classList={{
                      "text-fuchsia-500 rounded-md bg-white dark:bg-black":
                        instance.dbIdentifier ===
                        state.instance_active?.dbIdentifier,
                    }}
                    onClick={() =>
                      batch(() => {
                        setState(
                          "instance_active",
                          JSON.parse(
                            JSON.stringify(state.instance_list[index()]),
                          ),
                        );
                        setSelectInstanceOpen(false);
                      })
                    }
                  >
                    <p>{instance.systemId}</p>
                    <InstanceDetails instance={instance} />
                  </li>
                )}
              </For>
            </ul>
          </Show>
        </section>
      </Show>
      {!state.instance_active ? (
        ""
      ) : (
        <InstanceDetails instance={state.instance_active} />
      )}
    </section>
  );

  function H1Header() {
    return (
      <h1 class="text-lg xs:text-2xl font-semibold truncate">
        {state.instance_active?.systemId || ""}
      </h1>
    );
  }

  function InstanceDetails(props: { instance: Instance }) {
    return (
      <p class="text-neutral-500 flex flex-wrap items-baseline gap-x-4 gap-y-1 dark:text-neutral-400 text-xs sm:text-sm">
        <span>{humanizeSystemType(props.instance.systemType)}</span>
        <span>{props.instance.systemScope}</span>
      </p>
    );
  }
}

function humanizeSystemType(systemType: string): string {
  switch (systemType) {
    case "amazon_rds":
      return "Amazon RDS";
    case "google_cloudsql":
      return "Google Cloud SQL";
    case "azure_database":
      return "Azure Database";
    case "heroku":
      return "Heroku";
    case "crunchy_bridge":
      return "Crunchy Bridge";
    case "aiven":
      return "Aiven";
    case "tembo":
      return "Tembo";
    case "self_hosted":
      return "Self-hosted";
    default:
      return systemType.replace(/_/g, " ");
  }
}
