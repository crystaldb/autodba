import { HiOutlineMoon, HiOutlineSun } from "solid-icons/hi";
import { createEffect, createSignal, Show } from "solid-js";

interface DarkmodeSelectorProps {
  class?: string;
}

export function DarkmodeSelector(props: DarkmodeSelectorProps) {
  const [open, setOpen] = createSignal(false);
  const [darkmode, setDarkmode] = createSignal(
    document.documentElement.classList.contains("dark") ? "dark" : "light",
  );
  createEffect(() => {
    switch (darkmode()) {
      case "dark":
        document.documentElement.classList.add("dark");
        localStorage.theme = "dark"; // Whenever the user explicitly chooses dark mode
        break;
      case "light":
        document.documentElement.classList.remove("dark");
        localStorage.theme = "light"; // Whenever the user explicitly chooses light mode
        break;
      default:
        localStorage.removeItem("theme"); // Whenever the user explicitly chooses to respect the OS preference

        if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
          document.documentElement.classList.add("dark");
        } else {
          document.documentElement.classList.remove("dark");
        }
        break;
    }
    setOpen(false);
  });

  return (
    <section class={`flex gap-x-0.5 ${props.class}`}>
      <button
        onClick={() => setOpen(!open())}
        type="button"
        class="relative inline-flex items-center rounded-md px-3 py-2 text-sm font-semibold ring-1 ring-inset ring-zinc-200 focus:z-10 dark:block dark:text-zinc-300 dark:hover:bg-zinc-900 hover:bg-zinc-50 dark:ring-zinc-800"
      >
        <HiOutlineSun size="24" class="dark:hidden" />
        <HiOutlineMoon
          size="24"
          class="hidden dark:block dark:text-zinc-300 dark:hover:bg-zinc-900 hover:bg-zinc-50 dark:ring-zinc-300"
        />
      </button>
      <Show when={open()}>
        <span class="isolate inline-flex rounded-md shadow-sm">
          <button
            onClick={() => setDarkmode("dark")}
            type="button"
            class="relative inline-flex items-center rounded-s-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300"
            classList={{
              "bg-gray-300": darkmode() === "dark",
              "bg-white hover:bg-gray-50 focus:z-10": darkmode() !== "dark",
            }}
          >
            <HiOutlineMoon size="24" />
          </button>
          <button
            onClick={() => setDarkmode("light")}
            type="button"
            class="-ms-px relative inline-flex items-center px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300"
            classList={{
              "bg-gray-300": darkmode() === "light",
              "bg-white hover:bg-gray-50 focus:z-10": darkmode() !== "light",
            }}
          >
            <HiOutlineSun size="24" />
          </button>
          <button
            onClick={() => setDarkmode("system")}
            type="button"
            class="-ms-px relative inline-flex items-center rounded-e-md px-3 py-2 text-2xs whitespace-pre leading-none font-semibold text-gray-900 ring-1 ring-inset ring-gray-300"
            classList={{
              "bg-gray-300": darkmode() === "system",
              "bg-white hover:bg-gray-50 focus:z-10": darkmode() !== "system",
            }}
          >
            System{"\n"}default
          </button>
        </span>
      </Show>
    </section>
  );
}
