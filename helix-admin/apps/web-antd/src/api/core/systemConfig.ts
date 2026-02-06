import { requestClient } from '#/api/request';

export namespace SystemConfigApi {
  /** 系统配置项 */
  export interface ConfigItem {
    key: string;
    value: string;
    description: string;
  }
}

/**
 * 获取所有系统配置
 */
export async function getSystemConfigs() {
  return requestClient.get<SystemConfigApi.ConfigItem[]>('/admin/system/configs', {
    responseReturn: 'body',
  });
}

/**
 * 更新系统配置
 * @param key - 配置键
 * @param value - 配置值
 */
export async function updateSystemConfig(key: string, value: string) {
  return requestClient.put<{ ok: boolean }>(
    `/admin/system/configs/${key}`,
    { value },
    { responseReturn: 'body' }
  );
}
