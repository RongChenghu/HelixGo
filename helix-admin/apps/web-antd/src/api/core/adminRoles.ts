import { requestClient } from '#/api/request';

export namespace AdminRolesApi {
  export interface Role {
    name: string;
    description?: string;
    perms?: string[];
  }

  // v0.4 简化为纯列表，保留对 { list: Role[] } 的兼容
  export type ListResult = Role[] | { list: Role[] };
}

/**
 * 获取所有角色定义
 */
export async function getAdminRoles() {
  const res = await requestClient.get<AdminRolesApi.ListResult>(
    '/admin/admin-roles',
    {
      responseReturn: 'body',
    },
  );

  if (Array.isArray(res)) {
    return res;
  }
  return res.list || [];
}

