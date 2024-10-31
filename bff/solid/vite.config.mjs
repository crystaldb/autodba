import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";
// import devtools from "solid-devtools/vite";

export default defineConfig({
  plugins: [
    /*
    Uncomment the following line to enable solid-devtools.
    For more info see https://github.com/thetarnav/solid-devtools/tree/main/packages/extension#readme
    */
    // devtools(),
    solidPlugin(),
  ],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:5001/",
        secure: false,
        headers: {
          "ACCESS_KEY": process.env.VITE_ACCESS_KEY || "DEFAULT-ACCESS-KEY",
        },
        //"pathRewrite": { "^/api": "" }
      },
    },
  },
  build: {
    target: "esnext",
  },
});
