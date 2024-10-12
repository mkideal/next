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
                  "Generate <b>code</b> in multiple languages <b>flexibly</b>, focus on your data <b>structures</b>.",
                description:
                  "Home page hero title, can contain simple html tags",
              }),
            }}
          />
        </Heading>
        <div className={styles.indexCtas}>
          <Link className="button button--primary" to="/docs">
            <Translate>Get Started</Translate>
          </Link>
          <Link
            className="button button--info download-button"
            to="/docs/download"
          >
            <Translate>Download</Translate>
          </Link>
        </div>
      </div>
    </div>
  );
}

function TopBanner() {
  const version = useDocusaurusContext().siteConfig.customFields
    ?.version as string;

  return (
    version && (
      <div className={styles.topBanner}>
        <div className={styles.topBannerTitle}>
          {"🎉\xa0"}
          <Link to={"/docs/api/latest"} className={styles.topBannerTitleText}>
            <Translate
              id="homepage.banner.launch.newVersion"
              values={{ newVersion: version }}
            >
              {"Next \xa0{newVersion} is\xa0out!️"}
            </Translate>
          </Link>
          {"\xa0🥳"}
        </div>
      </div>
    )
  );
}

export default function Home(): JSX.Element {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title}`}
      description="Next is a powerful Generic Interface Definition Language(IDL) designed to create highly customized code across multiple programming languages."
    >
      <TopBanner />
      <HeroBanner />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
