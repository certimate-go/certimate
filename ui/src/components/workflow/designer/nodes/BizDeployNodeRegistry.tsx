import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconPackage } from "@tabler/icons-react";
import { Avatar } from "antd";

import { deploymentProvidersMap } from "@/domain/provider";
import { newNode } from "@/domain/workflow";

import { getAllPreviousNodes } from "../_util";
import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizDeployNodeConfigForm from "../forms/BizDeployNodeConfigForm";

export const BizDeployNodeRegistry: NodeRegistry = {
  type: NodeType.BizDeploy,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.deploy.label"),

    icon: IconPackage,
    iconColor: "#fff",
    iconBgColor: "#5b65f5",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizDeployNodeConfigForm.getSchema({}).safeParse(value);
        if (!res.success) {
          return {
            message: res.error.message,
            level: FeedbackLevel.Error,
          };
        }
      },
      ["config.certificateOutputNodeId"]: ({ value, context: { node } }) => {
        if (value == null) return;

        const prevNodeIds = getAllPreviousNodes(node).map((e) => e.id);
        if (!prevNodeIds.includes(value)) {
          return {
            message: "Invalid input",
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
              <Field<string> name="config.provider">
                {({ field: { value } }) => (
                  <>
                    {value ? (
                      <>
                        <div className="flex-1 truncate">{t(deploymentProvidersMap.get(value)?.name ?? "")}</div>
                        <Avatar shape="square" src={deploymentProvidersMap.get(value)?.icon} size={20} />
                      </>
                    ) : (
                      t("workflow.detail.design.editor.placeholder")
                    )}
                  </>
                )}
              </Field>
            </div>
          }
        />
      );
    },
  },

  onAdd: () => {
    return newNode(NodeType.BizDeploy, { i18n: getI18n() });
  },
};
