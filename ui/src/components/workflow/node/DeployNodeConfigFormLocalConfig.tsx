import { useTranslation } from "react-i18next";
import { IconChevronDown } from "@tabler/icons-react";
import { Alert, Button, Dropdown, Form, type FormInstance, Input, Select } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import CodeInput from "@/components/CodeInput";
import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

type DeployNodeConfigFormLocalConfigFieldValues = Nullish<{
  format: string;
  certPath: string;
  certPathForServerOnly?: string;
  certPathForIntermediaOnly?: string;
  keyPath?: string;
  pfxPassword?: string;
  jksAlias?: string;
  jksKeypass?: string;
  jksStorepass?: string;
  shellEnv?: string;
  preCommand?: string;
  postCommand?: string;
}>;

export type DeployNodeConfigFormLocalConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormLocalConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormLocalConfigFieldValues) => void;
};

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const SHELLENV_SH = "sh" as const;
const SHELLENV_CMD = "cmd" as const;
const SHELLENV_POWERSHELL = "powershell" as const;

const initFormModel = (): DeployNodeConfigFormLocalConfigFieldValues => {
  return {
    format: FORMAT_PEM,
    certPath: "/etc/ssl/certimate/cert.crt",
    keyPath: "/etc/ssl/certimate/cert.key",
    shellEnv: SHELLENV_SH,
  };
};

export const initPresetScript = (
  key: "sh_backup_files" | "ps_backup_files" | "sh_reload_nginx" | "ps_binding_iis" | "ps_binding_netsh" | "ps_binding_rdp",
  params?: {
    certPath?: string;
    certPathForServerOnly?: string;
    certPathForIntermediaOnly?: string;
    keyPath?: string;
    pfxPassword?: string;
    jksAlias?: string;
    jksKeypass?: string;
    jksStorepass?: string;
  }
) => {
  switch (key) {
    case "sh_backup_files":
      return `# 请将以下路径替换为实际值
cp "${params?.certPath || "<your-cert-path>"}" "${params?.certPath || "<your-cert-path>"}.bak" 2>/dev/null || :
cp "${params?.keyPath || "<your-key-path>"}" "${params?.keyPath || "<your-key-path>"}.bak" 2>/dev/null || :
      `.trim();

    case "ps_backup_files":
      return `# 请将以下路径替换为实际值
if (Test-Path -Path "${params?.certPath || "<your-cert-path>"}" -PathType Leaf) {
  Copy-Item -Path "${params?.certPath || "<your-cert-path>"}" -Destination "${params?.certPath || "<your-cert-path>"}.bak" -Force
}
if (Test-Path -Path "${params?.keyPath || "<your-key-path>"}" -PathType Leaf) {
  Copy-Item -Path "${params?.keyPath || "<your-key-path>"}" -Destination "${params?.keyPath || "<your-key-path>"}.bak" -Force
}
      `.trim();

    case "sh_reload_nginx":
      return `# *** 需要 root 权限 ***

sudo service nginx reload
      `.trim();

    case "ps_binding_iis":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "${params?.certPath || "<your-cert-path>"}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "${params?.pfxPassword || "<your-pfx-password>"}" # PFX 密码（与表单中保持一致）
$siteName = "<your-site-name>" # IIS 网站名称
$domain = "<your-domain-name>" # 域名
$ipaddr = "<your-binding-ip>"  # 绑定 IP，“*”表示所有 IP 绑定
$port = "<your-binding-port>"  # 绑定端口

# 导入证书到本地计算机的个人存储区
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 导入 WebAdministration 模块
Import-Module WebAdministration
# 检查是否已存在 HTTPS 绑定
$existingBinding = Get-WebBinding -Name "$siteName" -Protocol "https" -Port $port -HostHeader "$domain" -ErrorAction SilentlyContinue
if (!$existingBinding) {
    # 添加新的 HTTPS 绑定
  New-WebBinding -Name "$siteName" -Protocol "https" -Port $port -IPAddress "$ipaddr" -HostHeader "$domain"
}
# 获取绑定对象
$binding = Get-WebBinding -Name "$siteName" -Protocol "https" -Port $port -IPAddress "$ipaddr" -HostHeader "$domain"
# 绑定 SSL 证书
$binding.AddSslCertificate($thumbprint, "My")
# 删除目录下的证书文件
Remove-Item -Path "$pfxPath" -Force
      `.trim();

    case "ps_binding_netsh":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "${params?.certPath || "<your-cert-path>"}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "${params?.pfxPassword || "<your-pfx-password>"}" # PFX 密码（与表单中保持一致）
$ipaddr = "<your-binding-ip>"  # 绑定 IP，“0.0.0.0”表示所有 IP 绑定，可填入域名
$port = "<your-binding-port>"  # 绑定端口

# 导入证书到本地计算机的个人存储区
$addr = $ipaddr + ":" + $port
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 检测端口是否绑定证书，如绑定则删除绑定
$isExist = netsh http show sslcert ipport=$addr
if ($isExist -like "*$addr*"){ netsh http delete sslcert ipport=$addr }
# 绑定到端口
netsh http add sslcert ipport=$addr certhash=$thumbprint
# 删除目录下的证书文件
Remove-Item -Path "$pfxPath" -Force
      `.trim();

    case "ps_binding_rdp":
      return `# *** 需要管理员权限 ***

# 请将以下变量替换为实际值
$pfxPath = "${params?.certPath || "<your-cert-path>"}" # PFX 文件路径（与表单中保持一致）
$pfxPassword = "${params?.pfxPassword || "<your-pfx-password>"}" # PFX 密码（与表单中保持一致）

# 导入证书到本地计算机的个人存储区
$cert = Import-PfxCertificate -FilePath "$pfxPath" -CertStoreLocation Cert:\\LocalMachine\\My -Password (ConvertTo-SecureString -String "$pfxPassword" -AsPlainText -Force) -Exportable
# 获取 Thumbprint
$thumbprint = $cert.Thumbprint
# 绑定到 RDP
$rdpCertPath = "HKLM:\\SYSTEM\\CurrentControlSet\\Control\\Terminal Server\\WinStations\\RDP-Tcp"
Set-ItemProperty -Path $rdpCertPath -Name "SSLCertificateSHA1Hash" -Value "$thumbprint"
      `.trim();
  }
};

const DeployNodeConfigFormLocalConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormLocalConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.local_format.placeholder")),
    certPath: z
      .string()
      .min(1, t("workflow_node.deploy.form.local_cert_path.tooltip"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    certPathForServerOnly: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    certPathForIntermediaOnly: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    keyPath: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_PEM || !!v?.trim(), t("workflow_node.deploy.form.local_key_path.tooltip")),
    pfxPassword: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_PFX || !!v?.trim(), t("workflow_node.deploy.form.local_pfx_password.tooltip")),
    jksAlias: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.local_jks_alias.tooltip")),
    jksKeypass: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.local_jks_keypass.tooltip")),
    jksStorepass: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.local_jks_storepass.tooltip")),
    shellEnv: z.literal([SHELLENV_SH, SHELLENV_CMD, SHELLENV_POWERSHELL], t("workflow_node.deploy.form.local_shell_env.placeholder")),
    preCommand: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish(),
    postCommand: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish(),
  });
  const formRule = createSchemaFieldRule(formSchema);

  const fieldFormat = Form.useWatch("format", formInst);
  const fieldCertPath = Form.useWatch("certPath", formInst);

  const handleFormatSelect = (value: string) => {
    if (fieldFormat === value) return;

    switch (value) {
      case FORMAT_PEM:
        {
          if (/(.pfx|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue("certPath", fieldCertPath.replace(/(.pfx|.jks)$/, ".crt"));
          }
        }
        break;

      case FORMAT_PFX:
        {
          if (/(.crt|.jks)$/.test(fieldCertPath)) {
            formInst.setFieldValue("certPath", fieldCertPath.replace(/(.crt|.jks)$/, ".pfx"));
          }
        }
        break;

      case FORMAT_JKS:
        {
          if (/(.crt|.pfx)$/.test(fieldCertPath)) {
            formInst.setFieldValue("certPath", fieldCertPath.replace(/(.crt|.pfx)$/, ".jks"));
          }
        }
        break;
    }
  };

  const handlePresetPreScriptClick = (key: string) => {
    switch (key) {
      case "sh_backup_files":
      case "ps_backup_files":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue("certPath"),
            keyPath: formInst.getFieldValue("keyPath"),
          };
          formInst.setFieldValue("shellEnv", SHELLENV_SH);
          formInst.setFieldValue("preCommand", initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  const handlePresetPostScriptClick = (key: string) => {
    switch (key) {
      case "sh_reload_nginx":
        {
          formInst.setFieldValue("shellEnv", SHELLENV_SH);
          formInst.setFieldValue("postCommand", initPresetScript(key));
        }
        break;

      case "ps_binding_iis":
      case "ps_binding_netsh":
      case "ps_binding_rdp":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue("certPath"),
            pfxPassword: formInst.getFieldValue("pfxPassword"),
          };
          formInst.setFieldValue("shellEnv", SHELLENV_POWERSHELL);
          formInst.setFieldValue("postCommand", initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  const handleFormChange = (_: unknown, values: z.infer<typeof formSchema>) => {
    onValuesChange?.(values);
  };

  return (
    <Form
      form={formInst}
      disabled={disabled}
      initialValues={initialValues ?? initFormModel()}
      layout="vertical"
      name={formName}
      onValuesChange={handleFormChange}
    >
      <Form.Item>
        <Alert type="info" message={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local.guide") }}></span>} />
      </Form.Item>

      <Form.Item name="format" label={t("workflow_node.deploy.form.local_format.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.local_format.placeholder")} onSelect={handleFormatSelect}>
          <Select.Option key={FORMAT_PEM} value={FORMAT_PEM}>
            {t("workflow_node.deploy.form.local_format.option.pem.label")}
          </Select.Option>
          <Select.Option key={FORMAT_PFX} value={FORMAT_PFX}>
            {t("workflow_node.deploy.form.local_format.option.pfx.label")}
          </Select.Option>
          <Select.Option key={FORMAT_JKS} value={FORMAT_JKS}>
            {t("workflow_node.deploy.form.local_format.option.jks.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="certPath"
        label={t("workflow_node.deploy.form.local_cert_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_cert_path.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.local_cert_path.placeholder")} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name="keyPath"
          label={t("workflow_node.deploy.form.local_key_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_key_path.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_key_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name="certPathForServerOnly"
          label={t("workflow_node.deploy.form.local_servercert_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_servercert_path.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.local_servercert_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name="certPathForIntermediaOnly"
          label={t("workflow_node.deploy.form.local_intermediacert_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_intermediacert_path.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.local_intermediacert_path.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          name="pfxPassword"
          label={t("workflow_node.deploy.form.local_pfx_password.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_pfx_password.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_pfx_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name="jksAlias"
          label={t("workflow_node.deploy.form.local_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          name="jksKeypass"
          label={t("workflow_node.deploy.form.local_jks_keypass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_keypass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_keypass.placeholder")} />
        </Form.Item>

        <Form.Item
          name="jksStorepass"
          label={t("workflow_node.deploy.form.local_jks_storepass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.local_jks_storepass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.local_jks_storepass.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item name="shellEnv" label={t("workflow_node.deploy.form.local_shell_env.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.local_shell_env.placeholder")}>
          <Select.Option key={SHELLENV_SH} value={SHELLENV_SH}>
            {t("workflow_node.deploy.form.local_shell_env.option.sh.label")}
          </Select.Option>
          <Select.Option key={SHELLENV_CMD} value={SHELLENV_CMD}>
            {t("workflow_node.deploy.form.local_shell_env.option.cmd.label")}
          </Select.Option>
          <Select.Option key={SHELLENV_POWERSHELL} value={SHELLENV_POWERSHELL}>
            {t("workflow_node.deploy.form.local_shell_env.option.powershell.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item noStyle>
        <label className="mb-1 block">
          <div className="flex w-full items-center justify-between gap-4">
            <div className="max-w-full grow truncate">
              <span>{t("workflow_node.deploy.form.local_pre_command.label")}</span>
            </div>
            <div className="text-right">
              <Dropdown
                menu={{
                  items: ["sh_backup_files", "ps_backup_files"].map((key) => ({
                    key,
                    label: t(`workflow_node.deploy.form.local_preset_scripts.option.${key}.label`),
                    onClick: () => handlePresetPreScriptClick(key),
                  })),
                }}
                trigger={["click"]}
              >
                <Button size="small" type="link">
                  {t("workflow_node.deploy.form.local_preset_scripts.button")}
                  <IconChevronDown size="1.25em" />
                </Button>
              </Dropdown>
            </div>
          </div>
        </label>
        <Form.Item name="preCommand" rules={[formRule]}>
          <CodeInput
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language={["shell", "powershell"]}
            placeholder={t("workflow_node.deploy.form.local_pre_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item noStyle>
        <label className="mb-1 block">
          <div className="flex w-full items-center justify-between gap-4">
            <div className="max-w-full grow truncate">
              <span>{t("workflow_node.deploy.form.local_post_command.label")}</span>
            </div>
            <div className="text-right">
              <Dropdown
                menu={{
                  items: ["sh_reload_nginx", "ps_binding_iis", "ps_binding_netsh", "ps_binding_rdp"].map((key) => ({
                    key,
                    label: t(`workflow_node.deploy.form.local_preset_scripts.option.${key}.label`),
                    onClick: () => handlePresetPostScriptClick(key),
                  })),
                }}
                trigger={["click"]}
              >
                <Button size="small" type="link">
                  {t("workflow_node.deploy.form.local_preset_scripts.button")}
                  <IconChevronDown size="1.25em" />
                </Button>
              </Dropdown>
            </div>
          </div>
        </label>
        <Form.Item name="postCommand" rules={[formRule]}>
          <CodeInput
            height="auto"
            minHeight="64px"
            maxHeight="256px"
            language={["shell", "powershell"]}
            placeholder={t("workflow_node.deploy.form.local_post_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormLocalConfig;
