import React from "react";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import { FaWindows, FaApple, FaLinux } from "react-icons/fa";
import styles from "./styles.module.css";

interface DownloadItem {
  title: string;
  description: string;
  fileName: string;
  icon: React.ComponentType<React.ComponentProps<"svg">>;
  chipIcon?: React.ComponentType<React.ComponentProps<"svg">>;
}

const FeaturedDownloads: React.FC = () => {
  const { siteConfig } = useDocusaurusContext();
  const version = siteConfig.customFields.version as string;
  const repo = siteConfig.customFields.repo as string;
  const downloadUrl = (fileName: string) =>
    `${repo}/releases/download/v${version}/${fileName}`;

  const downloads: DownloadItem[] = [
    {
      title: "Microsoft Windows",
      description: "Windows 10 or later, Intel 64-bit processor",
      fileName: `next${version}.windows-amd64.msi`,
      icon: FaWindows,
    },
    {
      title: "Apple macOS (ARM64)",
      description: "macOS 11 or later, Apple 64-bit processor",
      fileName: `next${version}.darwin-arm64.tar.gz`,
      icon: FaApple,
    },
    {
      title: "Apple macOS (x86-64)",
      description: "macOS 11 or later, Intel 64-bit processor",
      fileName: `next${version}.darwin-amd64.tar.gz`,
      icon: FaApple,
    },
    {
      title: "Linux",
      description: "Linux 2.6.32 or later, Intel 64-bit processor",
      fileName: `next${version}.linux-amd64.tar.gz`,
      icon: FaLinux,
    },
  ];

  return (
    <div className={styles.featuredDownloads}>
      <div className="row">
        {downloads.map((item, index) => (
          <div key={index} className="col col--6 margin-bottom--md">
            <a
              href={downloadUrl(item.fileName)}
              className={styles.downloadItemLink}
            >
              <div className={styles.downloadItem}>
                <div>
                  <h4 className={styles.downloadTitle}>
                    <item.icon className={styles.downloadIcon} />
                    <span>{item.title}</span>
                  </h4>
                  <p>{item.description}</p>
                  <span className={styles.downloadFileName}>
                    {item.fileName}
                  </span>
                </div>
              </div>
            </a>
          </div>
        ))}
      </div>
    </div>
  );
};

export default FeaturedDownloads;
