import clsx from "clsx";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<"svg">>;
  description: JSX.Element;
};

const FeatureList: FeatureItem[] = [
  {
    title: "Multi-Language Code Generation",
    Svg: require("@site/static/img/feature_multilang.svg").default,
    description: (
      <>
        Generate code for multiple programming languages from a single source
        definition. This feature allows developers to define constants, enums,
        structures, and interfaces in a language-agnostic way. It ensures
        consistency across different programming environments and reduces the
        need for manual code maintenance in multiple languages.
      </>
    ),
  },
  {
    title: "Flexible customization",
    Svg: require("@site/static/img/feature_customization.svg").default,
    description: (
      <>
        Use a powerful template system that supports inheritance and overloading
        for flexible code output. Developers can create or customize templates
        for target languages, modifying the code generation process to meet
        specific project requirements without starting from scratch.
      </>
    ),
  },
  {
    title: "Annotation support",
    Svg: require("@site/static/img/feature_annotations.svg").default,
    description: (
      <>
        Provide additional metadata or instructions for code generation through
        annotations. Annotations allow developers to embed language-specific
        details or control the behavior of the code generator, enabling
        fine-grained customization and enhancing the flexibility of the
        generated code.
      </>
    ),
  },
];

function Feature({ title, Svg, description }: FeatureItem) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): JSX.Element {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
