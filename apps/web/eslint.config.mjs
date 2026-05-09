import nextConfig from "eslint-config-next";
import prettierConfig from "eslint-config-prettier";

export default [
  {
    ignores: [
      ".next/*",
      ".velite/*",
      "node_modules/*",
      "public/sw.js",
      "ui/*",
      "**/*.d.ts",
      "**/*.config.*"
    ],
  },
  ...nextConfig,
  prettierConfig,
  {
    rules: {
      // Custom rules
      "@next/next/no-html-link-for-pages": "off"
    }
  }
];
