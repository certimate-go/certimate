import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconContract } from "@tabler/icons-react";
import { Avatar } from "antd";

import { acmeDns01ProvidersMap, acmeHttp01ProvidersMap } from "@/domain/provider";
import { newNode } from "@/domain/workflow";

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

      type MapValueType<M> = M extends Map<string, infer V> ? V : never;
      const acmeProvidersMap = new Map<string, MapValueType<typeof acmeDns01ProvidersMap | typeof acmeHttp01ProvidersMap>>([
        ...acmeDns01ProvidersMap,
        ...acmeHttp01ProvidersMap,
      ]);

      return (
        <BaseNode
          description={
            <div className="flex items-center justify-between gap-1">
              <Field<string> name="config.domains">
                {({ field: { value } }) => {
                  return <div className="flex-1 truncate">{value || t("workflow.detail.design.editor.placeholder")}</div>;
                }}
              </Field>
              <Field<string> name="config.provider">
                {({ field: { value } }) => (value ? <Avatar shape="square" src={acmeProvidersMap.get(value)?.icon} size={20} /> : <></>)}
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
