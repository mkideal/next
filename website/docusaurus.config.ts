import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";
import type { Options as ClientRedirectsOptions } from "@docusaurus/plugin-client-redirects";

import prismLight from "./src/utils/prismLight";
import prismDark from "./src/utils/prismDark";

const URL = "https://next.as";
const ORGANIZATION_NAME = "mkideal";
const PROJECT_NAME = "next";
const REPO = `https://github.com/${ORGANIZATION_NAME}/${PROJECT_NAME}`;

const customFields = {
  version: "0.2.3",
  repo: REPO,
};

const config: Config = {
  title: "Next",
  tagline: "A generic IDL for generating customizable code across languages.",
  favicon: "img/favicon.ico",

  // Set the production url of your site here
  url: URL,
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: "/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: ORGANIZATION_NAME, // Usually your GitHub org/user name.
  projectName: PROJECT_NAME, // Usually your repo name.
  trailingSlash: false,

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  markdown: {
    mermaid: true,
  },
  themes: ["@docusaurus/theme-mermaid"],
  plugins: [
    [
      "client-redirects",
      {
        fromExtensions: ["html"],
        createRedirects(routePath) {
          // Redirect to /docs from /docs/get-started (now docs root doc)
          if (routePath === "/docs" || routePath === "/docs/") {
            return [`${routePath}/get-started`];
          }
          return [];
        },
      } satisfies ClientRedirectsOptions,
    ],
  ],

  customFields: customFields,

  presets: [
    [
      "classic",
      {
        docs: {
          path: "docs",
          sidebarPath: "./sidebars.ts",
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          // editUrl: REPO + "/tree/main/website/",
          // showLastUpdateAuthor: true,
          showLastUpdateTime: true,
        },
        //blog: {
        //  path: "blog",
        //  showReadingTime: true,
        //  feedOptions: {
        //    type: ["rss", "atom"],
        //    xslt: true,
        //  },
        //  // Please change this to your repo.
        //  // Remove this to remove the "edit this page" links.
        //  editUrl: REPO + "/tree/main/website/",
        //  // Useful options to enforce blogging best practices
        //  onInlineTags: "warn",
        //  onInlineAuthors: "warn",
        //  onUntruncatedBlogPosts: "warn",
        //  showLastUpdateAuthor: true,
        //  showLastUpdateTime: true,
        //},
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: "img/docusaurus-social-card.jpg",
    docs: {
      sidebar: {
        hideable: true,
        autoCollapseCategories: true,
      },
    },
    tableOfContents: {
      minHeadingLevel: 2,
      maxHeadingLevel: 5,
    },
    navbar: {
      title: "Next",
      logo: {
        alt: "Next Logo",
        src: "img/logo.png",
      },
      items: [
        // Left links
        {
          type: "doc",
          position: "left",
          docId: "get-started",
          label: "Docs",
        },
        {
          type: "docSidebar",
          position: "left",
          sidebarId: "api",
          label: "API",
        },
        // TODO: add showcase
        //{
        //  type: "docSidebar",
        //  position: "left",
        //  sidebarId: "showcase",
        //  label: "Showcase",
        //},
        // Right links
        {
          href: REPO,
          position: "right",
          className: "header-github-link",
          "aria-label": "GitHub repository",
        },
        // TODO: add algolia search
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          label: "Feature Images Designed by Freepik",
          href: "https://www.freepik.com",
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Next, Inc. Built with Docusaurus.`,
    },
    prism: {
      theme: prismLight,
      darkTheme: prismDark,
      additionalLanguages: [
        "protobuf",
        "java",
        "c",
        "cpp",
        "csharp",
        "bash",
        "go",
      ],
      magicComments: [
        {
          className: "theme-code-block-highlighted-line",
          line: "highlight-next-line",
          block: { start: "highlight-start", end: "highlight-end" },
        },
        {
          className: "code-block-error-line",
          line: "This will error",
        },
      ],
    },
    algolia: {
      // The application ID provided by Algolia
      appId: "FTUFY3O9F2",

      // Public API key: it is safe to commit it
      apiKey: "00e8a16f9517f447382d41ca398884a7",

      // The search index name
      indexName: "prod_next_doc",

      //contextualSearch: true,

      // Optional: Specify domains where the navigation should occur through window.location
      // instead on history.push. Useful when our Algolia config crawls multiple documentation
      // sites and we want to navigate with window.location.href to them.
      //externalUrlRegex: "external\\.com|domain\\.com",

      // Optional: Replace parts of the item URLs from Algolia. Useful when using the same search
      // index for multiple deployments using a different baseUrl. You can use regexp or string
      // in the `from` param. For example: localhost:3000 vs myCompany.com/docs
      //replaceSearchResultPathname: {
      //  from: "/docs/", // or as RegExp: /\/docs\//
      //  to: "/",
      //},

      // Optional: Algolia search parameters
      //searchParameters: {},

      // Optional: path for search page that enabled by default (`false` to disable it)
      //searchPagePath: "search",

      // Optional: whether the insights feature is enabled or not on Docsearch (`false` by default)
      //insights: false,

      //... other Algolia params
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
