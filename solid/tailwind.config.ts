import type { Config } from "tailwindcss";

import { green } from "tailwindcss/colors";
// @ts-expect-error
import pluginRtl from "tailwindcss-rtl";
import pluginTypography from "@tailwindcss/typography";

const config: Config = {
  content: ["./src/index.html", "./src/**/*.{js,jsx,ts,tsx}"],
  darkMode: "class",
  theme: {
    extend: {
      screens: {
        xs: "512px",
      },
      colors: {
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
