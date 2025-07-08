import { useTranslation } from "react-i18next";
import { DownOutlined as DownOutlinedIcon } from "@ant-design/icons";
import { Button, Dropdown, Form, type FormInstance, Input, Select, Switch } from "antd";
import { createSchemaFieldRule } from "antd-zod";
import { z } from "zod/v4";

import CodeInput from "@/components/CodeInput";
import Show from "@/components/Show";
import { CERTIFICATE_FORMATS } from "@/domain/certificate";

import { initPresetScript as _initPresetScript } from "./DeployNodeConfigFormLocalConfig";

type DeployNodeConfigFormSSHConfigFieldValues = Nullish<{
  format: string;
  certPath: string;
  certPathForServerOnly?: string;
  certPathForIntermediaOnly?: string;
  keyPath?: string;
  pfxPassword?: string;
  jksAlias?: string;
  jksKeypass?: string;
  jksStorepass?: string;
  preCommand?: string;
  postCommand?: string;
  useSCP?: boolean;
}>;

export type DeployNodeConfigFormSSHConfigProps = {
  form: FormInstance;
  formName: string;
  disabled?: boolean;
  initialValues?: DeployNodeConfigFormSSHConfigFieldValues;
  onValuesChange?: (values: DeployNodeConfigFormSSHConfigFieldValues) => void;
};

const FORMAT_PEM = CERTIFICATE_FORMATS.PEM;
const FORMAT_PFX = CERTIFICATE_FORMATS.PFX;
const FORMAT_JKS = CERTIFICATE_FORMATS.JKS;

const initFormModel = (): DeployNodeConfigFormSSHConfigFieldValues => {
  return {
    format: FORMAT_PEM,
    certPath: "/etc/ssl/certimate/cert.crt",
    keyPath: "/etc/ssl/certimate/cert.key",
  };
};

const initPresetScript = (
  key: Parameters<typeof _initPresetScript>[0] | "sh_replace_synologydsm_ssl" | "sh_replace_fnos_ssl" | "sh_replace_qnap_ssl",
  params?: Parameters<typeof _initPresetScript>[1]
) => {
  switch (key) {
    case "sh_replace_synologydsm_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS
# 脚本参考 https://github.com/catchdave/ssl-certs/blob/main/replace_synology_ssl_certs.sh

# 请将以下变量替换为实际值
$tmpFullchainPath = "${params?.certPath || "<your-fullchain-cert-path>"}" # 证书文件路径（与表单中保持一致）
$tmpCertPath = "${params?.certPathForServerOnly || "<your-server-cert-path>"}" # 服务器证书文件路径（与表单中保持一致）
$tmpKeyPath = "${params?.keyPath || "<your-key-path>"}" # 私钥文件路径（与表单中保持一致）

DEBUG=1
error_exit() { echo "[ERROR] $1"; exit 1; }
warn() { echo "[WARN] $1"; }
info() { echo "[INFO] $1"; }
debug() { [[ "\${DEBUG}" ]] && echo "[DEBUG] $1"; }

certs_src_dir="/usr/syno/etc/certificate/system/default"
target_cert_dirs=(
  "/usr/syno/etc/certificate/system/FQDN"
  "/usr/local/etc/certificate/ScsiTarget/pkg-scsi-plugin-server/"
  "/usr/local/etc/certificate/SynologyDrive/SynologyDrive/"
  "/usr/local/etc/certificate/WebDAVServer/webdav/"
  "/usr/local/etc/certificate/ActiveBackup/ActiveBackup/"
  "/usr/syno/etc/certificate/smbftpd/ftpd/")

# 获取证书目录
default_dir_name=$(</usr/syno/etc/certificate/_archive/DEFAULT)
if [[ -n "$default_dir_name" ]]; then
  target_cert_dirs+=("/usr/syno/etc/certificate/_archive/\${default_dir_name}")
  debug "Default cert directory found: '/usr/syno/etc/certificate/_archive/\${default_dir_name}'"
else
  warn "No default directory found. Probably unusual? Check: 'cat /usr/syno/etc/certificate/_archive/DEFAULT'"
fi

# 获取反向代理证书目录
for proxy in /usr/syno/etc/certificate/ReverseProxy/*/; do
  debug "Found proxy dir: \${proxy}"
  target_cert_dirs+=("\${proxy}")
done

[[ "\${DEBUG}" ]] && set -x

# 复制文件
cp -rf "$tmpFullchainPath" "\${certs_src_dir}/fullchain.pem" || error_exit "Halting because of error moving fullchain file"
cp -rf "$tmpCertPath" "\${certs_src_dir}/cert.pem" || error_exit "Halting because of error moving cert file"
cp -rf "$tmpKeyPath" "\${certs_src_dir}/privkey.pem" || error_exit "Halting because of error moving privkey file"
chown root:root "\${certs_src_dir}/"{privkey,fullchain,cert}.pem || error_exit "Halting because of error chowning files"
info "Certs moved from /tmp & chowned."

# 替换证书
for target_dir in "\${target_cert_dirs[@]}"; do
  if [[ ! -d "$target_dir" ]]; then
    debug "Target cert directory '$target_dir' not found, skipping..."
    continue
  fi
  info "Copying certificates to '$target_dir'"
  if ! (cp "\${certs_src_dir}/"{privkey,fullchain,cert}.pem "$target_dir/" && \
    chown root:root "$target_dir/"{privkey,fullchain,cert}.pem); then
      warn "Error copying or chowning certs to \${target_dir}"
  fi
done

# 重启服务
info "Rebooting all the things..."
/usr/syno/bin/synosystemctl restart nmbd
/usr/syno/bin/synosystemctl restart avahi
/usr/syno/bin/synosystemctl restart ldap-server
/usr/syno/bin/synopkg is_onoff ScsiTarget 1>/dev/null && /usr/syno/bin/synopkg restart ScsiTarget
/usr/syno/bin/synopkg is_onoff SynologyDrive 1>/dev/null && /usr/syno/bin/synopkg restart SynologyDrive
/usr/syno/bin/synopkg is_onoff WebDAVServer 1>/dev/null && /usr/syno/bin/synopkg restart WebDAVServer
/usr/syno/bin/synopkg is_onoff ActiveBackup 1>/dev/null && /usr/syno/bin/synopkg restart ActiveBackup
if ! /usr/syno/bin/synow3tool --gen-all && sudo /usr/syno/bin/synosystemctl restart nginx; then
  warn "nginx failed to restart"
fi

info "Completed"
      `.trim();

    case "sh_replace_fnos_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS
# 脚本参考 https://github.com/lfgyx/fnos_certificate_update/blob/main/src/update_cert.sh

# 请将以下变量替换为实际值
# 飞牛证书实际存放路径请在 \`/usr/trim/etc/network_cert_all.conf\` 中查看，注意不要修改文件名
$tmpFullchainPath = "${params?.certPath || "<your-fullchain-cert-path>"}" # 证书文件路径（与表单中保持一致）
$tmpCertPath = "${params?.certPathForServerOnly || "<your-server-cert-path>"}" # 服务器证书文件路径（与表单中保持一致）
$tmpKeyPath = "${params?.keyPath || "<your-key-path>"}" # 私钥文件路径（与表单中保持一致）
$fnFullchainPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/fullchain.crt" # 飞牛证书文件路径
$fnCertPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/example.com.crt" # 飞牛服务器证书文件路径
$fnKeyPath = "/usr/trim/var/trim_connect/ssls/example.com/1234567890/example.com.key" # 飞牛私钥文件路径
$domain = "<your-domain-name>" # 域名

# 复制文件
cp -rf "$tmpFullchainPath" "$fnFullchainPath"
cp -rf "$tmpCertPath" "$fnCertPath"
cp -rf "$tmpKeyPath" "$fnKeyPath"
chmod 755 "$fnFullchainPath"
chmod 755 "$fnCertPath"
chmod 755 "$fnKeyPath"

# 更新数据库
NEW_EFFECT_DATE=$(openssl x509 -startdate -noout -in "$fnCertPath" | sed "s/^.*=\\(.*\\)$/\\1/")
NEW_EFFECT_TIMESTAMP=$(date -d "$NEW_EFFECT_DATE" +%s%3N)
NEW_EXPIRY_DATE=$(openssl x509 -enddate -noout -in "$fnCertPath" | sed "s/^.*=\\(.*\\)$/\\1/")
NEW_EXPIRY_TIMESTAMP=$(date -d "$NEW_EXPIRY_DATE" +%s%3N)
psql -U postgres -d trim_connect -c "UPDATE cert SET valid_from=$NEW_EFFECT_TIMESTAMP, valid_to=$NEW_EXPIRY_TIMESTAMP WHERE domain='$domain'"

# 重启服务
systemctl restart webdav.service
systemctl restart smbftpd.service
systemctl restart trim_nginx.service
      `.trim();

    case "sh_replace_qnap_ssl":
      return `# *** 需要 root 权限 ***
# 注意仅支持替换证书，需本身已开启过一次 HTTPS

# 请将以下变量替换为实际值
$tmpFullchainPath = "${params?.certPath || "<your-fullchain-cert-path>"}" # 证书文件路径（与表单中保持一致）
$tmpKeyPath = "${params?.keyPath || "<your-key-path>"}" # 私钥文件路径（与表单中保持一致）

# 复制文件
cp -rf "$tmpFullchainPath" /etc/stunnel/backup.cert
cp -rf "$tmpKeyPath" /etc/stunnel/backup.key
cat /etc/stunnel/backup.key > /etc/stunnel/stunnel.pem
cat /etc/stunnel/backup.cert >> /etc/stunnel/stunnel.pem
chmod 600 /etc/stunnel/backup.cert
chmod 600 /etc/stunnel/backup.key
chmod 600 /etc/stunnel/stunnel.pem

# 重启服务
/etc/init.d/stunnel.sh restart
/etc/init.d/reverse_proxy.sh reload
      `.trim();
  }

  return _initPresetScript(key as Parameters<typeof _initPresetScript>[0], params);
};

const DeployNodeConfigFormSSHConfig = ({ form: formInst, formName, disabled, initialValues, onValuesChange }: DeployNodeConfigFormSSHConfigProps) => {
  const { t } = useTranslation();

  const formSchema = z.object({
    format: z.literal([FORMAT_PEM, FORMAT_PFX, FORMAT_JKS], t("workflow_node.deploy.form.ssh_format.placeholder")),
    certPath: z
      .string()
      .min(1, t("workflow_node.deploy.form.ssh_cert_path.tooltip"))
      .max(256, t("common.errmsg.string_max", { max: 256 })),
    keyPath: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_PEM || !!v?.trim(), t("workflow_node.deploy.form.ssh_key_path.tooltip")),
    certPathForServerOnly: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    certPathForIntermediaOnly: z
      .string()
      .max(256, t("common.errmsg.string_max", { max: 256 }))
      .nullish(),
    pfxPassword: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_PFX || !!v?.trim(), t("workflow_node.deploy.form.ssh_pfx_password.tooltip")),
    jksAlias: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.ssh_jks_alias.tooltip")),
    jksKeypass: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.ssh_jks_keypass.tooltip")),
    jksStorepass: z
      .string()
      .max(64, t("common.errmsg.string_max", { max: 256 }))
      .nullish()
      .refine((v) => fieldFormat !== FORMAT_JKS || !!v?.trim(), t("workflow_node.deploy.form.ssh_jks_storepass.tooltip")),
    preCommand: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish(),
    postCommand: z
      .string()
      .max(20480, t("common.errmsg.string_max", { max: 20480 }))
      .nullish(),
    useSCP: z.boolean().nullish(),
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
          formInst.setFieldValue("preCommand", initPresetScript(key, presetScriptParams));
        }
        break;
    }
  };

  const handlePresetPostScriptClick = (key: string) => {
    switch (key) {
      case "sh_reload_nginx":
        {
          formInst.setFieldValue("postCommand", initPresetScript(key));
        }
        break;

      case "sh_replace_synologydsm_ssl":
      case "sh_replace_fnos_ssl":
      case "sh_replace_qnap_ssl":
        {
          const presetScriptParams = {
            certPath: formInst.getFieldValue("certPath"),
            certPathForServerOnly: formInst.getFieldValue("certPathForServerOnly"),
            certPathForIntermediaOnly: formInst.getFieldValue("certPathForIntermediaOnly"),
            keyPath: formInst.getFieldValue("keyPath"),
          };
          formInst.setFieldValue("postCommand", initPresetScript(key, presetScriptParams));
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
      <Form.Item name="format" label={t("workflow_node.deploy.form.ssh_format.label")} rules={[formRule]}>
        <Select placeholder={t("workflow_node.deploy.form.ssh_format.placeholder")} onSelect={handleFormatSelect}>
          <Select.Option key={FORMAT_PEM} value={FORMAT_PEM}>
            {t("workflow_node.deploy.form.ssh_format.option.pem.label")}
          </Select.Option>
          <Select.Option key={FORMAT_PFX} value={FORMAT_PFX}>
            {t("workflow_node.deploy.form.ssh_format.option.pfx.label")}
          </Select.Option>
          <Select.Option key={FORMAT_JKS} value={FORMAT_JKS}>
            {t("workflow_node.deploy.form.ssh_format.option.jks.label")}
          </Select.Option>
        </Select>
      </Form.Item>

      <Form.Item
        name="certPath"
        label={t("workflow_node.deploy.form.ssh_cert_path.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_cert_path.tooltip") }}></span>}
      >
        <Input placeholder={t("workflow_node.deploy.form.ssh_cert_path.placeholder")} />
      </Form.Item>

      <Show when={fieldFormat === FORMAT_PEM}>
        <Form.Item
          name="keyPath"
          label={t("workflow_node.deploy.form.ssh_key_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_key_path.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_key_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name="certPathForServerOnly"
          label={t("workflow_node.deploy.form.ssh_servercert_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_servercert_path.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ssh_servercert_path.placeholder")} />
        </Form.Item>

        <Form.Item
          name="certPathForIntermediaOnly"
          label={t("workflow_node.deploy.form.ssh_intermediacert_path.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_intermediacert_path.tooltip") }}></span>}
        >
          <Input allowClear placeholder={t("workflow_node.deploy.form.ssh_intermediacert_path.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_PFX}>
        <Form.Item
          name="pfxPassword"
          label={t("workflow_node.deploy.form.ssh_pfx_password.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_pfx_password.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_pfx_password.placeholder")} />
        </Form.Item>
      </Show>

      <Show when={fieldFormat === FORMAT_JKS}>
        <Form.Item
          name="jksAlias"
          label={t("workflow_node.deploy.form.ssh_jks_alias.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_alias.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_alias.placeholder")} />
        </Form.Item>

        <Form.Item
          name="jksKeypass"
          label={t("workflow_node.deploy.form.ssh_jks_keypass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_keypass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_keypass.placeholder")} />
        </Form.Item>

        <Form.Item
          name="jksStorepass"
          label={t("workflow_node.deploy.form.ssh_jks_storepass.label")}
          rules={[formRule]}
          tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_jks_storepass.tooltip") }}></span>}
        >
          <Input placeholder={t("workflow_node.deploy.form.ssh_jks_storepass.placeholder")} />
        </Form.Item>
      </Show>

      <Form.Item className="mb-0" htmlFor="null">
        <label className="mb-1 block">
          <div className="flex w-full items-center justify-between gap-4">
            <div className="max-w-full grow truncate">
              <span>{t("workflow_node.deploy.form.ssh_pre_command.label")}</span>
            </div>
            <div className="text-right">
              <Dropdown
                menu={{
                  items: ["sh_backup_files", "ps_backup_files"].map((key) => ({
                    key,
                    label: t(`workflow_node.deploy.form.ssh_preset_scripts.option.${key}.label`),
                    onClick: () => handlePresetPreScriptClick(key),
                  })),
                }}
                trigger={["click"]}
              >
                <Button size="small" type="link">
                  {t("workflow_node.deploy.form.ssh_preset_scripts.button")}
                  <DownOutlinedIcon />
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
            placeholder={t("workflow_node.deploy.form.ssh_pre_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item className="mb-0" htmlFor="null">
        <label className="mb-1 block">
          <div className="flex w-full items-center justify-between gap-4">
            <div className="max-w-full grow truncate">
              <span>{t("workflow_node.deploy.form.ssh_post_command.label")}</span>
            </div>
            <div className="text-right">
              <Dropdown
                menu={{
                  items: [
                    "sh_reload_nginx",
                    "sh_replace_synologydsm_ssl",
                    "sh_replace_fnos_ssl",
                    "sh_replace_qnap_ssl",
                    "ps_binding_iis",
                    "ps_binding_netsh",
                    "ps_binding_rdp",
                  ].map((key) => ({
                    key,
                    label: t(`workflow_node.deploy.form.ssh_preset_scripts.option.${key}.label`),
                    onClick: () => handlePresetPostScriptClick(key),
                  })),
                }}
                trigger={["click"]}
              >
                <Button size="small" type="link">
                  {t("workflow_node.deploy.form.ssh_preset_scripts.button")}
                  <DownOutlinedIcon />
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
            placeholder={t("workflow_node.deploy.form.ssh_post_command.placeholder")}
          />
        </Form.Item>
      </Form.Item>

      <Form.Item
        name="useSCP"
        label={t("workflow_node.deploy.form.ssh_use_scp.label")}
        rules={[formRule]}
        tooltip={<span dangerouslySetInnerHTML={{ __html: t("workflow_node.deploy.form.ssh_use_scp.tooltip") }}></span>}
      >
        <Switch />
      </Form.Item>
    </Form>
  );
};

export default DeployNodeConfigFormSSHConfig;
