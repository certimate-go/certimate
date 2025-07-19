import eslint from "@eslint/js";
import { defineConfig } from "eslint/config";
import tailwindcssPlugin from "eslint-plugin-better-tailwindcss";
import importPlugin from "eslint-plugin-import";
import prettierPluginConfig from "eslint-plugin-prettier/recommended";
import reactHooksPlugin from "eslint-plugin-react-hooks";
import reactRefreshPlugin from "eslint-plugin-react-refresh";
import typescriptPlugin from "typescript-eslint";

/**
 * @type {import("eslint").Linter.Config[]}
 */
export default defineConfig(
  // Basic
  eslint.configs["recommended"],
  {
    name: "eslint/import",
    extends: [importPlugin.flatConfigs["recommended"], importPlugin.flatConfigs["typescript"]],
    rules: {
      "import/no-named-as-default-member": "off",
      "import/no-unresolved": "off",
      "import/order": [
        "error",
        {
          groups: ["builtin", "external", "internal", ["parent", "sibling"], "index"],
          pathGroups: [
            {
              pattern: "react*",
              group: "external",
              position: "before",
            },
            {
              pattern: "react/**",
              group: "external",
              position: "before",
            },
            {
              pattern: "react-*",
              group: "external",
              position: "before",
            },
            {
              pattern: "react-*/**",
              group: "external",
              position: "before",
            },
            {
              pattern: "~/**",
              group: "external",
              position: "after",
            },
            {
              pattern: "@/**",
              group: "internal",
              position: "before",
            },
          ],
          pathGroupsExcludedImportTypes: ["builtin"],
          alphabetize: {
            order: "asc",
            caseInsensitive: true,
          },
        },
      ],
      "sort-imports": [
        "error",
        {
          ignoreDeclarationSort: true,
        },
      ],
    },
    settings: {
      "import/resolver": {
        node: {
          extensions: [".js", ".jsx", ".ts", ".tsx"],
        },
        typescript: {
          alwaysTryTypes: true,
        },
      },
    },
  },

  // Typescript
  {
    name: "typescript",
    extends: [typescriptPlugin.configs["recommended"]],
    rules: {
      "@typescript-eslint/consistent-type-imports": "error",
      "@typescript-eslint/no-empty-object-type": [
        "error",
        {
          allowInterfaces: "with-single-extends",
        },
      ],
      "@typescript-eslint/no-explicit-any": [
        "warn",
        {
          ignoreRestArgs: true,
        },
      ],
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
          destructuredArrayIgnorePattern: "^_",
          varsIgnorePattern: "^_",
        },
      ],
    },
  },

  // Pretter
  {
    name: "prettier",
    extends: [prettierPluginConfig],
  },

  // React
  {
    name: "react",
    extends: [reactHooksPlugin.configs["recommended-latest"], reactRefreshPlugin.configs["vite"]],
    rules: {
      "react-refresh/only-export-components": [
        "warn",
        {
          allowConstantExport: true,
        },
      ],
    },
  },

  // TailwindCSS
  {
    name: "tailwindcss",
    plugins: {
      "better-tailwindcss": tailwindcssPlugin,
    },
    rules: {
      ...tailwindcssPlugin.configs["recommended-warn"].rules,
      ...tailwindcssPlugin.configs["recommended-error"].rules,

      "better-tailwindcss/enforce-consistent-line-wrapping": "off",
      "better-tailwindcss/no-unregistered-classes": "off",
    },
    settings: {
      "better-tailwindcss": {
        entryPoint: "src/global.css",
      },
    },
  }
);
