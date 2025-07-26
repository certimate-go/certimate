import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { Outlet, useLocation, useNavigate, useParams } from "react-router-dom";
import { IconChevronDown, IconHistory, IconRobot, IconTrash } from "@tabler/icons-react";
import { App, Button, Dropdown, Flex, Form, Input, Segmented, Skeleton } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod";

import ModalForm from "@/components/ModalForm";
import Show from "@/components/Show";
import { isAllNodesValidated } from "@/domain/workflow";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { remove as removeWorkflow } from "@/repository/workflow";
import { useWorkflowStore } from "@/stores/workflow";
import { getErrMsg } from "@/utils/error";

const WorkflowDetail = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const { t } = useTranslation();

  const { message, modal, notification } = App.useApp();

  const { id: workflowId } = useParams();
  const { workflow, initialized, ...workflowState } = useWorkflowStore(useZustandShallowSelector(["workflow", "initialized", "init", "destroy", "setEnabled"]));
  useEffect(() => {
    workflowState.init(workflowId!);

    return () => {
      workflowState.destroy();
    };
  }, [workflowId]);

  const tabs = [
    ["design", "workflow.detail.design.tab", <IconRobot size="1em" />],
    ["runs", "workflow.detail.runs.tab", <IconHistory size="1em" />],
  ] satisfies [string, string, React.ReactElement][];
  const [tabValue, setTabValue] = useState<string>();
  useEffect(() => {
    const subpath = location.pathname.split("/")[3];
    if (!subpath) {
      navigate(`/workflows/${workflowId}/${tabs[0][0]}`, { replace: true });
      return;
    }

    setTabValue(subpath);
  }, [location.pathname, workflowId]);

  const handleTabChange = (value: string) => {
    setTabValue(value);
    navigate(`/workflows/${workflowId}/${value}`);

    workflowState.init(workflow.id); // reload state
  };

  const handleEnableClick = async () => {
    if (!workflow.enabled && (!workflow.content || !isAllNodesValidated(workflow.content))) {
      message.warning(t("workflow.action.enable.errmsg.uncompleted"));
      return;
    }

    try {
      await workflowState.setEnabled(!workflow.enabled);
    } catch (err) {
      console.error(err);
      notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
    }
  };

  const handleDeleteClick = () => {
    modal.confirm({
      title: <span className="text-error">{t("workflow.action.delete.modal.title", { name: workflow.name })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("workflow.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          const resp = await removeWorkflow(workflow);
          if (resp) {
            navigate("/workflows", { replace: true });
          }
        } catch (err) {
          console.error(err);
          notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  return (
    <div className="flex size-full flex-col">
      <div className="px-6 py-4">
        <div className="relative mx-auto max-w-320">
          <div className="flex justify-between gap-2">
            <div>
              <h1>{workflow.name || "\u00A0"}</h1>
              <p className="mb-0 text-base text-gray-500">{workflow.description || "\u00A0"}</p>
            </div>
            <Flex className="my-2" gap="small">
              {initialized
                ? [
                    <WorkflowBaseInfoModal key="edit" trigger={<Button>{t("common.button.edit")}</Button>} />,
                    <Button key="enable" onClick={handleEnableClick}>
                      {workflow.enabled ? t("workflow.action.disable.button") : t("workflow.action.enable.button")}
                    </Button>,
                    <Dropdown
                      key="more"
                      menu={{
                        items: [
                          {
                            key: "delete",
                            label: t("workflow.action.delete.button"),
                            danger: true,
                            icon: <IconTrash size="1.25em" />,
                            onClick: () => {
                              handleDeleteClick();
                            },
                          },
                        ],
                      }}
                      trigger={["click"]}
                    >
                      <Button icon={<IconChevronDown size="1.25em" />} iconPosition="end">
                        {t("common.button.more")}
                      </Button>
                    </Dropdown>,
                  ]
                : []}
            </Flex>
          </div>

          <div className="absolute -bottom-12 left-1/2 z-1 -translate-x-1/2">
            <Segmented
              className="shadow"
              options={tabs.map(([key, label, icon]) => ({
                value: key,
                label: <span className="px-2 text-sm">{t(label)}</span>,
                icon: (
                  <span className="anticon scale-125" role="img">
                    {icon}
                  </span>
                ),
              }))}
              size="large"
              value={tabValue}
              defaultValue="orchestration"
              onChange={handleTabChange}
            />
          </div>
        </div>
      </div>

      <div className="p-4">
        <Show
          when={initialized}
          fallback={
            <div className="mx-auto max-w-320 pt-12">
              <Skeleton active />
            </div>
          }
        >
          <Outlet />
        </Show>
      </div>
    </div>
  );
};

const WorkflowBaseInfoModal = ({ trigger }: { trigger?: React.ReactNode }) => {
  const { t } = useTranslation();

  const { notification } = App.useApp();

  const { workflow, ...workflowState } = useWorkflowStore(useZustandShallowSelector(["workflow", "setBaseInfo"]));

  const formSchema = z.object({
    name: z
      .string(t("workflow.detail.baseinfo.form.name.placeholder"))
      .min(1, t("workflow.detail.baseinfo.form.name.placeholder"))
      .max(64, t("common.errmsg.string_max", { max: 64 })),
    description: z
      .string(t("workflow.detail.baseinfo.form.description.placeholder"))
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);
  const {
    form: formInst,
    formPending,
    formProps,
    submit: submitForm,
  } = useAntdForm<z.infer<typeof formSchema>>({
    initialValues: { name: workflow.name, description: workflow.description },
    onSubmit: async (values) => {
      try {
        await workflowState.setBaseInfo(values.name!, values.description!);
      } catch (err) {
        notification.error({ message: t("common.text.request_error"), description: getErrMsg(err) });

        throw err;
      }
    },
  });

  const handleFormFinish = async () => {
    return submitForm();
  };

  return (
    <>
      <ModalForm
        disabled={formPending}
        layout="vertical"
        form={formInst}
        modalProps={{ destroyOnHidden: true }}
        okText={t("common.button.save")}
        title={t(`workflow.detail.baseinfo.modal.title`)}
        trigger={trigger}
        width={480}
        {...formProps}
        onFinish={handleFormFinish}
      >
        <Form.Item name="name" label={t("workflow.detail.baseinfo.form.name.label")} rules={[formRule]}>
          <Input placeholder={t("workflow.detail.baseinfo.form.name.placeholder")} />
        </Form.Item>

        <Form.Item name="description" label={t("workflow.detail.baseinfo.form.description.label")} rules={[formRule]}>
          <Input placeholder={t("workflow.detail.baseinfo.form.description.placeholder")} />
        </Form.Item>
      </ModalForm>
    </>
  );
};

export default WorkflowDetail;
