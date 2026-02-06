import { requestClient } from '#/api/request';

export namespace AdminAuditApi {
  export interface AuditLog {
    id: number;
    adminUserId: number;
    adminUsername: string | null;
    action: string;
    targetType: string | null;
    targetId: string | null;
    payload: any;
    ip: string | null;
    userAgent: string | null;
    createdAt: string;
  }

  export interface ListParams {
    page?: number;
    pageSize?: number;
    keyword?: string;
    action?: string;
    from?: string;
    to?: string;
  }

  export interface ListResult {
    list: AuditLog[];
    total: number;
    page: number;
    pageSize: number;
  }
}

/**
 * 获取审计日志列表
 */
export async function getAuditLogs(params: AdminAuditApi.ListParams) {
  return requestClient.get<AdminAuditApi.ListResult>('/admin/audit/logs', {
    params,
    responseReturn: 'body',
  });
}

