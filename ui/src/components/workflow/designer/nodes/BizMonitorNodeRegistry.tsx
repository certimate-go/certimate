import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconDeviceDesktopSearch } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizMonitorNodeConfigForm from "../forms/BizMonitorNodeConfigForm";

export const BizMonitorNodeRegistry: NodeRegistry = {
  type: NodeType.BizMonitor,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.monitor.label"),

    icon: IconDeviceDesktopSearch,
    iconColor: "#fff",
    iconBgColor: "#5b65f5",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizMonitorNodeConfigForm.getSchema({}).safeParse(value);
        if (!res.success) {
          return {
            message: res.error.message,
            level: FeedbackLevel.Error,
          };
        }
      },
    },

    render: () => {
      const { t } = getI18n();

      return (
        <BaseNode
          description={
            <Field name="config.domain">
              {({ field: { value: fieldDomain } }) => (
                <Field name="config.host">
                  {({ field: { value: fieldHost } }) => (
                    <>{fieldDomain || fieldHost ? fieldDomain || fieldHost : t("workflow.detail.design.editor.placeholder")}</>
                  )}
                </Field>
              )}
            </Field>
          }
        />
      );
    },
  },

  onAdd: () => {
    return newNode(NodeType.BizMonitor, { i18n: getI18n() });
  },
};
