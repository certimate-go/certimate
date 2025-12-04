import { ClientResponseError } from "pocketbase";

import { CA_PROVIDERS } from "@/domain/provider";
import {
  type EmailsSettingsContent,
  type NotifyTemplateContent,
  type PersistenceSettingsContent,
  SETTINGS_NAMES,
  type SSLProviderSettingsContent,
  type ScriptTemplateContent,
  type SettingsModel,
  type SettingsNames,
} from "@/domain/settings";

import { COLLECTION_NAME_SETTINGS, getPocketBase } from "./_pocketbase";

interface SettingsContentMap {
  [SETTINGS_NAMES.EMAILS]: EmailsSettingsContent;
  [SETTINGS_NAMES.NOTIFY_TEMPLATE]: NotifyTemplateContent;
  [SETTINGS_NAMES.SCRIPT_TEMPLATE]: ScriptTemplateContent;
  [SETTINGS_NAMES.SSL_PROVIDER]: SSLProviderSettingsContent;
  [SETTINGS_NAMES.PERSISTENCE]: PersistenceSettingsContent;
}

export const get = async <K extends SettingsNames | string, T extends NonNullable<unknown>>(
  name: K
): Promise<K extends keyof SettingsContentMap ? SettingsModel<SettingsContentMap[K]> : SettingsModel<T>> => {
  let resp: K extends keyof SettingsContentMap ? SettingsModel<SettingsContentMap[K]> : SettingsModel<T>;
  try {
    resp = await getPocketBase().collection(COLLECTION_NAME_SETTINGS).getFirstListItem<typeof resp>(`name='${name}'`, {
      requestKey: null,
    });
    return resp;
  } catch (err) {
    if (err instanceof ClientResponseError && err.status === 404) {
      resp = {
        name: name,
        content: {},
      } as unknown as typeof resp;
    } else {
      throw err;
    }
  }

  // 兜底设置一些默认值（需确保与后端默认值保持一致），防止视图层空指针
  switch (name) {
    case SETTINGS_NAMES.EMAILS:
      {
        resp.content ??= {};
        (resp.content as EmailsSettingsContent).emails ??= [];
      }
      break;

    case SETTINGS_NAMES.NOTIFY_TEMPLATE:
      {
        resp.content ??= {};
        (resp.content as NotifyTemplateContent).templates ??= [];
      }
      break;

    case SETTINGS_NAMES.SCRIPT_TEMPLATE:
      {
        resp.content ??= {};
        (resp.content as ScriptTemplateContent).templates ??= [];
      }
      break;

    case SETTINGS_NAMES.SSL_PROVIDER:
      {
        resp.content ??= {};
        (resp.content as SSLProviderSettingsContent).provider ??= CA_PROVIDERS.LETSENCRYPT;
      }
      break;

    case SETTINGS_NAMES.PERSISTENCE:
      {
        resp.content ??= {};
        (resp.content as PersistenceSettingsContent).certificatesWarningDaysBeforeExpire ??= 21;
        (resp.content as PersistenceSettingsContent).certificatesRetentionMaxDays ??= 0;
        (resp.content as PersistenceSettingsContent).workflowRunsRetentionMaxDays ??= 0;
      }
      break;
  }

  return resp;
};

export const save = async <T extends NonNullable<unknown>>(record: MaybeModelRecordWithId<SettingsModel<T>>) => {
  if (record.id) {
    return await getPocketBase().collection(COLLECTION_NAME_SETTINGS).update<SettingsModel<T>>(record.id, record);
  }

  return await getPocketBase().collection(COLLECTION_NAME_SETTINGS).create<SettingsModel<T>>(record);
};
