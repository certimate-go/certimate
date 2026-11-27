import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router-dom";
import { Tabs } from "antd";

import Show from "@/components/Show";

import PresetListNotifyTemplates from "./PresetListNotifyTemplates";
import PresetListScriptTemplates from "./PresetListScriptTemplates";
import PresetListWorkflowTemplates from "./PresetListWorkflowTemplates";

type PresetUsages = "notification" | "script" | "workflow";

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
      <div className="container mb-6">
        <h1 className="text-2xl font-medium text-foreground/80">{t("preset.page.heading")}</h1>
        <p className="mt-1 text-sm text-foreground/70">{t("preset.page.subheading")}</p>
      </div>

      <div className="container">
        <Tabs
          className="-mt-2"
          classNames={{
            header: "px-[2px]",
          }}
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
            // TODO: 暂时隐藏工作流模板入口
            // {
            //   key: "workflow",
            //   label: t("preset.props.usage.workflow"),
            // },
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
            <Show.Case when={tabKey === "workflow"}>
              <PresetListWorkflowTemplates />
            </Show.Case>
          </Show>
        </div>
      </div>
    </div>
  );
};

export default PresetList;
