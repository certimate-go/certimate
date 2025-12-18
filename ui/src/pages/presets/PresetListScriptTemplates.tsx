import { useState } from "react";
import { useTranslation } from "react-i18next";
import { IconDots, IconEdit, IconPlus, IconTrash } from "@tabler/icons-react";
import { useControllableValue, useMount } from "ahooks";
import { App, Button, Card, Dropdown, Form, Input, Typography } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { nanoid } from "nanoid/non-secure";
import { ClientResponseError } from "pocketbase";
import { z } from "zod";

import CodeTextInput from "@/components/CodeTextInput";
import DrawerForm from "@/components/DrawerForm";
import Tips from "@/components/Tips";
import { useAntdForm, useZustandShallowSelector } from "@/hooks";
import { useScriptTemplatesStore } from "@/stores/settings";
import { getErrMsg } from "@/utils/error";

const MAX_TEMPLATE_COUNT = 99;

type PresetTemplate = {
  name: string;
  command: string;
};

const PresetListScriptTemplates = () => {
  const { t } = useTranslation();

  const { message, modal, notification } = App.useApp();

  const { templates, loading, loadedAtOnce, fetchTemplates, setTemplates, addTemplate, removeTemplateByIndex } = useScriptTemplatesStore();
  useMount(() => {
    fetchTemplates().catch((err) => {
      if (err instanceof ClientResponseError && err.isAbort) {
        return;
      }

      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
    });
  });

  const [createDrawerOpen, setCreateDrawerOpen] = useState(false);
  const [detailDrawerOpen, setDetailDrawerOpen] = useState(false);
  const [detailDrawerRecord, setDetailDrawerRecord] = useState<PresetTemplate>();
  const [detailDrawerIndex, setDetailDrawerIndex] = useState<number>();

  const handleCreateClick = () => {
    if (!loadedAtOnce) return;

    if (templates.length >= MAX_TEMPLATE_COUNT) {
      message.warning(t("preset.warning.excceeded"));
      return;
    }

    setCreateDrawerOpen(true);
  };

  const handleRecordDetailClick = (template: PresetTemplate, index: number) => {
    setDetailDrawerIndex(index);
    setDetailDrawerRecord({ ...template });
    setDetailDrawerOpen(true);
  };

  const handleRecordDeleteClick = (template: PresetTemplate, index: number) => {
    modal.confirm({
      title: <span className="text-error">{t("preset.action.delete.modal.title", { name: template.name })}</span>,
      content: <span dangerouslySetInnerHTML={{ __html: t("preset.action.delete.modal.content") }} />,
      icon: (
        <span className="anticon" role="img">
          <IconTrash className="text-error" size="1em" />
        </span>
      ),
      okText: t("common.button.confirm"),
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await removeTemplateByIndex(index);
        } catch (err) {
          console.error(err);
          notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
        }
      },
    });
  };

  const handleCreateDrawerSubmit = async (values: PresetTemplate) => {
    try {
      await addTemplate(values);

      setCreateDrawerOpen(false);
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
    }
  };

  const handleModifyDrawerSubmit = async (values: PresetTemplate) => {
    try {
      const newTemplates = [...templates];
      newTemplates[detailDrawerIndex!] = values;
      await setTemplates(newTemplates);

      setDetailDrawerIndex(void 0);
      setDetailDrawerRecord(void 0);
      setDetailDrawerOpen(false);
    } catch (err) {
      console.error(err);
      notification.error({ title: t("common.text.request_error"), description: getErrMsg(err) });
    }
  };

  return (
    <>
      <Tips className="mb-4" message={<span dangerouslySetInnerHTML={{ __html: t("preset.props.usage.script.tips") }}></span>} />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
        <div className="h-40">
          <Card
            className="size-full text-gray-500 transition-all select-none hover:text-stone-800 dark:hover:text-stone-200"
            styles={{
              root: {
                height: "100%",
              },
              body: {
                height: "100%",
              },
            }}
            hoverable
            onClick={handleCreateClick}
          >
            <div className="flex size-full flex-col items-center justify-center gap-4 py-4">
              <IconPlus size={36} stroke="1.25" />
              <div>{t("preset.action.create.button")}</div>
            </div>
          </Card>
        </div>

        {templates.map((template, index) => (
          <div className="h-40">
            <Card
              key={template.name}
              className="size-full"
              styles={{
                body: {
                  height: "100%",
                  padding: "1rem",
                },
                header: {
                  padding: "0.5rem 1rem",
                },
              }}
              extra={
                <Dropdown
                  menu={{
                    items: [
                      {
                        key: "edit",
                        label: t("preset.action.modify.menu"),
                        icon: (
                          <span className="anticon scale-125">
                            <IconEdit size="1em" />
                          </span>
                        ),
                        onClick: (e) => {
                          e.domEvent.stopPropagation();
                          handleRecordDetailClick(template, index);
                        },
                      },
                      {
                        type: "divider",
                      },
                      {
                        key: "delete",
                        label: t("preset.action.delete.menu"),
                        danger: true,
                        icon: (
                          <span className="anticon scale-125">
                            <IconTrash size="1em" />
                          </span>
                        ),
                        onClick: (e) => {
                          e.domEvent.stopPropagation();
                          handleRecordDeleteClick(template, index);
                        },
                      },
                    ],
                  }}
                  trigger={["click"]}
                >
                  <Button
                    icon={<IconDots size="1.25em" />}
                    type="text"
                    onClick={(e) => {
                      e.stopPropagation();
                    }}
                  />
                </Dropdown>
              }
              hoverable
              title={<Typography.Text ellipsis>{template.name}</Typography.Text>}
              onClick={() => {
                handleRecordDetailClick(template, index);
              }}
            >
              <Typography.Paragraph className="whitespace-pre-line" ellipsis={{ rows: 3 }} type="secondary">
                {template.command}
              </Typography.Paragraph>
            </Card>
          </div>
        ))}

        {loading && !loadedAtOnce && (
          <div className="h-40">
            <Card className="size-full" loading size="small" />
          </div>
        )}
      </div>

      <InternalEditDrawer
        data={{ name: "", command: "" }}
        mode={"create"}
        open={createDrawerOpen}
        afterClose={() => setCreateDrawerOpen(false)}
        onOpenChange={(open) => setCreateDrawerOpen(open)}
        onSubmit={handleCreateDrawerSubmit}
      />
      <InternalEditDrawer
        data={detailDrawerRecord}
        mode={"modify"}
        open={detailDrawerOpen}
        afterClose={() => setDetailDrawerOpen(false)}
        onOpenChange={(open) => setDetailDrawerOpen(open)}
        onSubmit={handleModifyDrawerSubmit}
      />
    </>
  );
};

const InternalEditDrawer = ({
  mode,
  data,
  onSubmit,
  ...props
}: {
  afterClose?: () => void;
  mode: "create" | "modify";
  data?: Nullish<PresetTemplate>;
  open: boolean;
  onOpenChange?: (open: boolean) => void;
  onSubmit?: (record: PresetTemplate) => void;
}) => {
  const { t } = useTranslation();

  const { templates } = useScriptTemplatesStore(useZustandShallowSelector(["templates"]));

  const [open, setOpen] = useControllableValue<boolean>(props, {
    valuePropName: "open",
    defaultValuePropName: "defaultOpen",
    trigger: "onOpenChange",
  });

  const afterClose = () => {
    formInst.resetFields();
    props.afterClose?.();
  };

  const formSchema = z
    .object({
      name: z.string().nonempty(t("preset.form.name.placeholder")),
      command: z.string().nonempty(t("preset.form.script_command.placeholder")),
    })
    .superRefine((values, ctx) => {
      if (values.name) {
        const name = String(values.name).trim();
        const duplicatedCount = templates.filter((t) => String(t.name).trim() === name).length;
        if (duplicatedCount > (mode === "create" ? 0 : 1)) {
          ctx.addIssue({
            code: "custom",
            message: t("preset.form.name.errmsg.duplicated"),
            path: ["name"],
          });
        }
      }
    });
  const formRule = createSchemaFieldRule(formSchema);
  const { form: formInst, formProps } = useAntdForm<z.infer<typeof formSchema>>({
    name: "viewPresetListScriptTemplates_InternalDrawerForm_" + nanoid(),
    initialValues: data,
  });

  const handleFormFinish = async (values: z.infer<typeof formSchema>) => {
    switch (mode) {
      case "create":
      case "modify":
        {
          onSubmit?.(values);
        }
        break;

      default:
        throw "Invalid props: `mode`";
    }

    setOpen(false);
  };

  return (
    <DrawerForm
      {...formProps}
      clearOnDestroy
      drawerProps={{ autoFocus: true, destroyOnHidden: true, size: "large", afterOpenChange: (open) => !open && afterClose?.() }}
      form={formInst}
      layout="vertical"
      okText={mode === "create" ? t("common.button.create") : mode === "modify" ? t("common.button.save") : void 0}
      open={open}
      preserve={false}
      title={mode === "create" ? t("preset.action.create.modal.title") : mode === "modify" ? t("preset.action.modify.modal.title") : void 0}
      validateTrigger="onSubmit"
      onFinish={handleFormFinish}
      onOpenChange={props.onOpenChange}
    >
      <Form.Item name="name" label={t("preset.form.name.label")} rules={[formRule]}>
        <Input maxLength={100} placeholder={t("preset.form.name.placeholder")} />
      </Form.Item>

      <Form.Item name="command" label={t("preset.form.script_command.label")} rules={[formRule]}>
        <CodeTextInput height="auto" minHeight="256px" language={["shell", "powershell"]} placeholder={t("preset.form.script_command.placeholder")} />
      </Form.Item>
    </DrawerForm>
  );
};

export default PresetListScriptTemplates;
