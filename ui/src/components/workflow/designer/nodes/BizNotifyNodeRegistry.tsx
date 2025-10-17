import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconSend } from "@tabler/icons-react";
import { Avatar } from "antd";

import { notificationProvidersMap } from "@/domain/provider";
import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizNotifyNodeConfigForm from "../forms/BizNotifyNodeConfigForm";

export const BizNotifyNodeRegistry: NodeRegistry = {
  type: NodeType.BizNotify,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.notify.label"),

    icon: IconSend,
    iconColor: "#fff",
    iconBgColor: "#0693d4",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizNotifyNodeConfigForm.getSchema({}).safeParse(value);
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
              <Field<string> name="config.provider">
                {({ field: { value } }) => (
                  <>
                    {value ? (
                      <>
                        <div className="flex-1 truncate">{t(notificationProvidersMap.get(value)?.name ?? "")}</div>
                        <Avatar shape="square" src={notificationProvidersMap.get(value)?.icon} size={20} />
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
    return newNode(NodeType.BizNotify, { i18n: getI18n() });
  },
};
