import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconHourglassHigh } from "@tabler/icons-react";

import { newNode } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import DelayNodeConfigForm from "../forms/DelayNodeConfigForm";

export const DelayNodeRegistry: NodeRegistry = {
  type: NodeType.Delay,

  kind: NodeKindType.Basis,

  meta: {
    labelText: getI18n().t("workflow_node.delay.label"),

    icon: IconHourglassHigh,
    iconColor: "#2a354c",
    iconBgColor: "#fed421",

    clickable: true,
    expandable: false,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = DelayNodeConfigForm.getSchema({}).safeParse(value);
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
            <Field name="config.wait">
              {({ field: { value } }) => (
                <>
                  <div>{value != null ? `${value} ${t("workflow_node.delay.form.wait.unit")}` : t("workflow.detail.design.editor.placeholder")}</div>
                </>
              )}
            </Field>
          }
        />
      );
    },
  },

  onAdd() {
    return newNode(NodeType.Delay, { i18n: getI18n() });
  },
};
