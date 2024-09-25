import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";
import type { Options as ClientRedirectsOptions } from "@docusaurus/plugin-client-redirects";

import Prism from "prismjs";
import prismLight from "./src/utils/prismLight";
import prismDark from "./src/utils/prismDark";

if (typeof window !== "undefined") {
  window.Prism = Prism;
} else if (typeof global !== "undefined") {
  global.Prism = Prism;
}

// FIXME: change to https://next.as
const URL = "https://gopherd.com";
// FIXME: change to "https://github.com/next/next"
const REPO = "https://github.com/gopherd/next";

const config: Config = {
  title: "Next",
  tagline: "A generic IDL for generating customizable code across languages.",
  favicon: "img/favicon.ico",

  // Set the production url of your site here
  url: URL,
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  // FIXME: change to "/"
  baseUrl: "/next/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  // FIXME: change to "next" when organization name "next" is available.
  organizationName: "gopherd", // Usually your GitHub org/user name.
  projectName: "next", // Usually your repo name.
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

  customFields: {
    announcedVersion: "0.0.3",
  },

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
        {
          type: "docSidebar",
          position: "left",
          sidebarId: "showcase",
          label: "Showcase",
        },
        //{ to: "/blog", label: "Blog", position: "left" },
        // Right links
        {
          href: REPO,
          position: "right",
          className: "header-github-link",
          "aria-label": "GitHub repository",
        },
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
  } satisfies Preset.ThemeConfig,
};

export default config;
