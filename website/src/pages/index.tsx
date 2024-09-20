import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import useBaseUrl, { useBaseUrlUtils } from "@docusaurus/useBaseUrl";
import Translate, { translate } from "@docusaurus/Translate";
import Layout from "@theme/Layout";
import HomepageFeatures from "@site/src/components/HomepageFeatures";
import Heading from "@theme/Heading";

import styles from "./index.module.css";

function HeroBanner() {
  return (
    <div className={styles.hero} data-theme="dark">
      <div className={styles.heroInner}>
        <Heading as="h1" className={styles.heroProjectTagline}>
          <span
            className={styles.heroTitleTextHtml}
            // eslint-disable-next-line react/no-danger
            dangerouslySetInnerHTML={{
              __html: translate({
                id: "homepage.hero.title",
                message:
                  "<b>Define</b> Data Types,<br/><b>Generate</b> Multi-Language Code,<br/><b>Customize</b> Fully.",
                description:
                  "Home page hero title, can contain simple html tags",
              }),
            }}
          />
        </Heading>
        <div className={styles.indexCtas}>
          <Link className="button button--info" to="/docs">
            <Translate>Get Started</Translate>
          </Link>
        </div>
      </div>
    </div>
  );
}

function TopBanner() {
  const announcedVersion = useDocusaurusContext().siteConfig.customFields
    ?.announcedVersion as string;

  return (
    <div className={styles.topBanner}>
      <div className={styles.topBannerTitle}>
        {"🎉\xa0"}
        <Link
          to={`/blog/releases/${announcedVersion}`}
          className={styles.topBannerTitleText}
        >
          <Translate
            id="homepage.banner.launch.newVersion"
            values={{ newVersion: announcedVersion }}
          >
            {"Next \xa0{newVersion} is\xa0out!️"}
          </Translate>
        </Link>
        {"\xa0🥳"}
      </div>
    </div>
  );
}

export default function Home(): JSX.Element {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`Hello from ${siteConfig.title}`}
      description="Description will go into a meta tag in <head />"
    >
      <TopBanner />
      <HeroBanner />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
