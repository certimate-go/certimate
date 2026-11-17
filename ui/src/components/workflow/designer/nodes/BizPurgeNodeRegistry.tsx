import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconBucketDroplet } from "@tabler/icons-react";
import { Avatar } from "antd";

import { getPurgeProvider } from "@/domain/purgeProvider";
import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import BizPurgeNodeConfigForm from "../forms/BizPurgeNodeConfigForm";

export const BizPurgeNodeRegistry: NodeRegistry = {
  type: NodeType.BizPurge,

  kind: NodeKindType.Business,

  meta: {
    labelText: getI18n().t("workflow_node.purge.label"),

    icon: IconBucketDroplet,
    iconColor: "#fff",
    iconBgColor: "#5b65f5",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = BizPurgeNodeConfigForm.getSchema({}).safeParse(value);
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
                        <div className="flex-1 truncate">{t(getPurgeProvider(value)?.name ?? value)}</div>
                        {getPurgeProvider(value) ? <Avatar shape="square" src={getPurgeProvider(value)?.icon} size={20} /> : null}
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
    return newNode(NodeType.BizPurge, { i18n: getI18n() });
  },
};
