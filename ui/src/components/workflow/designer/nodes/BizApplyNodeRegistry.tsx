import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconContract } from "@tabler/icons-react";
import { Avatar } from "antd";

import { acmeHttp01ProvidersMap, getACMEDns01Provider } from "@/domain/provider";
import { type WorkflowNodeConfigForBizApply, newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizApplyNodeConfigForm from "../forms/BizApplyNodeConfigForm";

export const BizApplyNodeRegistry: NodeRegistry = {
  type: NodeType.BizApply,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.apply.label"),

    icon: IconContract,
    iconColor: "#fff",
    iconBgColor: "#5b65f5",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizApplyNodeConfigForm.getSchema({}).safeParse(value);
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
            <div className="flex items-center justify-between gap-1">
              <Field<WorkflowNodeConfigForBizApply> name="config">
                {({ field: { value } }) => {
                  const displayText = value.identifier === "domain" ? value.domains : value.identifier === "ip" ? value.ipaddrs : void 0;
                  return <div className="flex-1 truncate">{displayText || t("workflow.detail.design.editor.placeholder")}</div>;
                }}
              </Field>
              <Field<string> name="config.provider">
                {({ field: { value } }) =>
                  value ? <Avatar shape="square" src={(getACMEDns01Provider(value) ?? acmeHttp01ProvidersMap.get(value))?.icon} size={20} /> : <></>
                }
              </Field>
            </div>
          }
        />
      );
    },
  },

  onAdd: () => {
    return newNode(NodeType.BizApply, { i18n: getI18n() });
  },
};
