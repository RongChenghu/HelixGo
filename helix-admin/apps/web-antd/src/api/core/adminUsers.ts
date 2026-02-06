import { requestClient } from '#/api/request';

export namespace AdminUsersApi {
  export interface AdminUser {
    id: number;
    username: string;
    isEnabled: boolean;
    roles: string[];
    createdAt?: string;
    updatedAt?: string;
    lastLoginAt?: string | null;
  }

  export interface ListParams {
    page?: number;
    pageSize?: number;
    keyword?: string;
  }

  // 后端 v0.4 目前返回的是简单列表，预留兼容分页包裹结构
  export interface ListResult {
    list: AdminUser[];
    total?: number;
    page?: number;
    pageSize?: number;
  }

  export interface CreateParams {
    username: string;
    password: string;
    roles?: string[];
  }

  export interface EnableParams {
    enabled: boolean;
  }

  export interface ResetPasswordParams {
    password: string;
  }

  export interface RolesResult {
    roles: string[];
  }

  export interface SetRolesParams {
    roles: string[];
  }
}

/**
 * 获取管理员用户列表
 */
export async function getAdminUsers(params: AdminUsersApi.ListParams) {
  const res = await requestClient.get<AdminUsersApi.AdminUser[] | AdminUsersApi.ListResult>(
    '/admin/admin-users',
    {
      params,
      responseReturn: 'body',
    },
  );

  if (Array.isArray(res)) {
    return {
      list: res,
      total: res.length,
      page: 1,
      pageSize: res.length,
    } as AdminUsersApi.ListResult;
  }

  return res;
}

/**
 * 创建管理员用户
 */
export async function createAdminUser(data: AdminUsersApi.CreateParams) {
  return requestClient.post<{ id: number; username: string }>(
    '/admin/admin-users',
    data,
    {
      responseReturn: 'body',
    },
  );
}

/**
 * 设置管理员启用/禁用状态
 */
export async function setAdminEnabled(id: number, isEnabled: boolean) {
  return requestClient.post<{ ok: boolean }>(
    `/admin/admin-users/${id}/enable`,
    {
      enabled: isEnabled,
    } as AdminUsersApi.EnableParams,
    {
      responseReturn: 'body',
    },
  );
}

/**
 * 重置管理员密码
 */
export async function resetAdminPassword(id: number, newPassword: string) {
  return requestClient.post<{ ok: boolean }>(
    `/admin/admin-users/${id}/reset-password`,
    {
      password: newPassword,
    } as AdminUsersApi.ResetPasswordParams,
    {
      responseReturn: 'body',
    },
  );
}

/**
 * 获取管理员角色列表
 */
export async function getAdminUserRoles(id: number) {
  return requestClient.get<AdminUsersApi.RolesResult>(`/admin/admin-users/${id}/roles`, {
    responseReturn: 'body',
  });
}

/**
 * 设置管理员角色
 */
export async function setAdminUserRoles(id: number, roles: string[]) {
  return requestClient.post<{ success: boolean }>(`/admin/admin-users/${id}/roles`, {
    roles,
  }, {
    responseReturn: 'body',
  });
}

