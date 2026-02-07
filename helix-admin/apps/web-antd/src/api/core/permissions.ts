import { requestClient } from '#/api/request';

export interface PermissionItem {
  code: string;
  description: string;
}

/**
 * 获取权限字典（用于角色新建/编辑时的权限多选）
 */
export async function getPermissions(): Promise<PermissionItem[]> {
  const res = await requestClient.get<PermissionItem[]>(
    '/admin/permissions',
    { responseReturn: 'body' },
  );
  return Array.isArray(res) ? res : [];
}
