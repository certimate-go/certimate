import { getI18n } from "react-i18next";
import { FeedbackLevel, Field } from "@flowgram.ai/fixed-layout-editor";
import { IconRocket } from "@tabler/icons-react";

import { WORKFLOW_TRIGGERS } from "@/domain/workflow";

import { BaseNode } from "./_shared";
import { NodeKindType, type NodeRegistry, NodeType } from "./typings";
import StartNodeConfigForm from "../forms/StartNodeConfigForm";

export const StartNodeRegistry: NodeRegistry = {
  type: NodeType.Start,

  kind: NodeKindType.Basis,

  meta: {
    labelText: getI18n().t("workflow_node.start.label"),

    icon: IconRocket,
    iconColor: "#fff",
    iconBgColor: "#ed6d0c",

    isStart: true,

    clickable: true,
    expandable: false,
    selectable: false,

    addDisable: true,
    copyDisable: true,
    deleteDisable: true,
  },

  formMeta: {
    validate: {
      ["config"]: ({ value }) => {
        const res = StartNodeConfigForm.getSchema({}).safeParse(value);
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
              <Field name="config.trigger">
                {({ field: { value: fieldTrigger } }) => (
                  <>
                    <div>
                      {fieldTrigger === WORKFLOW_TRIGGERS.SCHEDULED
                        ? t("workflow.props.trigger.scheduled")
                        : fieldTrigger === WORKFLOW_TRIGGERS.MANUAL
                          ? t("workflow.props.trigger.manual")
                          : t("workflow.detail.design.editor.placeholder")}
                    </div>
                    <div>
                      <Field name="config.triggerCron">
                        {({ field: { value: fieldTriggerCron } }) => <>{fieldTrigger === WORKFLOW_TRIGGERS.SCHEDULED ? fieldTriggerCron || "\u00A0" : ""}</>}
                      </Field>
                    </div>
                  </>
                )}
              </Field>
            </div>
          }
        />
      );
    },
  },

  canAdd: () => {
    return false;
  },

  canDelete: () => {
    return false;
  },
};
