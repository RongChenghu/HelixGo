import { requestClient } from '#/api/request';

export namespace AdminRolesApi {
  export interface Role {
    id: number;
    name: string;
    description?: string;
    perms?: string[];
  }

  export type ListResult = Role[] | { list: Role[] };

  export interface CreateBody {
    name: string;
    description?: string;
    perms?: string[];
  }

  export interface UpdateBody {
    description?: string;
    perms?: string[];
  }
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

/**
 * 新建角色
 */
export async function createRole(body: AdminRolesApi.CreateBody) {
  return requestClient.post<{ id: number; name: string }>(
    '/admin/admin-roles',
    body,
    { responseReturn: 'body' },
  );
}

/**
 * 更新角色（描述与权限）
 */
export async function updateRole(
  id: number,
  body: AdminRolesApi.UpdateBody,
) {
  return requestClient.put<{ id: number; name: string }>(
    `/admin/admin-roles/${id}`,
    body,
    { responseReturn: 'body' },
  );
}

/**
 * 删除角色（内置 admin 角色不可删）
 */
export async function deleteRole(id: number) {
  return requestClient.delete<{ id: number }>(`/admin/admin-roles/${id}`, {
    responseReturn: 'body',
  });
}

