import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { Tabs } from "antd";

import Show from "@/components/Show";

import PresetListNotifyTemplates from "./PresetListNotifyTemplates";
import PresetListScriptTemplates from "./PresetListScriptTemplates";

type PresetUsages = "notification" | "script";

const PresetList = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const { t } = useTranslation();

  const [tabKey, setTabKey] = useState<PresetUsages>(() => {
    return (searchParams.get("usage") || "notification") as PresetUsages;
  });

  const handleTabChange = (key: string) => {
    setTabKey(key as PresetUsages);
    setSearchParams((prev) => {
      prev.set("usage", key);
      return prev;
    });
  };

  return (
    <div className="px-6 py-4">
      <div className="container">
        <h1>{t("preset.page.title")}</h1>
        <p className="text-base text-gray-500">{t("preset.page.subtitle")}</p>
      </div>

      <div className="container">
        <Tabs
          className="-mt-2"
          activeKey={tabKey}
          items={[
            {
              key: "notification",
              label: t("preset.props.usage.notification"),
            },
            {
              key: "script",
              label: t("preset.props.usage.script"),
            },
          ]}
          size="large"
          onChange={(key) => handleTabChange(key)}
        />

        <div className="relative">
          <Show>
            <Show.Case when={tabKey === "notification"}>
              <PresetListNotifyTemplates />
            </Show.Case>
            <Show.Case when={tabKey === "script"}>
              <PresetListScriptTemplates />
            </Show.Case>
          </Show>
        </div>
      </div>
    </div>
  );
};

export default PresetList;
