import { defineConfig, globalIgnores } from "eslint/config";
import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default defineConfig([
  js.configs.recommended,
  ...tseslint.configs.recommended,

  globalIgnores(["node_modules/**", ".next/**", "out/**", "build/**", "dist/**"]),

  {
    rules: {
      /**
       * JavaScript
       */
      "no-unreachable": "error",
      "no-dupe-keys": "error",
      "no-constant-condition": "error",

      /**
       * Best Practices
       */
      "no-console": [
        "warn",
        {
          allow: ["warn", "error"],
        },
      ],
    },
  },
]);
