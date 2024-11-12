import type { Config } from "tailwindcss";

import pluginTypography from "@tailwindcss/typography";
// @ts-expect-error tailwindcss-rtl has no typescript definition
import pluginRtl from "tailwindcss-rtl";
import { green, zinc } from "tailwindcss/colors";

const config: Config = {
  content: ["./src/index.html", "./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      screens: {
        xs: "512px",
      },
      colors: {
        backgrounddark: zinc[900],
        backgroundlite: zinc[100],
        notificationdark: green,
        notificationlite: green,
      },
      fontSize: {
        "3xs": ["0.5rem", "0.625rem"],
        "2xs": ["0.625rem", "0.75rem"],
      },
      spacing: {
        128: "32rem",
      },
    },
  },
  plugins: [pluginRtl, pluginTypography],
};

export default config;
