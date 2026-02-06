import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    accessToken: string;
  }

  export interface RefreshTokenResult {
    data: string;
    status: number;
  }
}

/**
 * 登录 - 对接 /admin/auth/login
 */
export async function loginApi(data: AuthApi.LoginParams) {
  // 后端 /admin/auth/login 返回形态：
  // 1) 直接裸 JSON：{ token, admin }
  // 2) 或经过默认拦截器包了一层：{ data: { token, admin }, ... }
  const res = await requestClient.post<any>(
    '/admin/auth/login',
    {
      username: data.username,
      password: data.password,
    },
    {
      responseReturn: 'body',
    },
  );

  let token: string | undefined;

  if (res && typeof res === 'object') {
    // 优先从顶层 token，其次从 data.token 中取
    token = res.token ?? res.data?.token;
  } else if (typeof res === 'string') {
    token = res;
  }

  // 兼容 Vben 期望的返回格式 { accessToken }
  return {
    accessToken: token || '',
  } as AuthApi.LoginResult;
}

/**
 * 获取当前管理员信息 - 对接 /admin/auth/me
 */
export async function getAdminMe() {
  // 后端直接返回用户信息对象，使用 body 模式
  const res = await requestClient.get<{
    id: string | number;
    name: string;
    roles: string[];
    permissions: string[];
  }>('/admin/auth/me', {
    responseReturn: 'body', // 使用 body 模式，直接返回响应体
  });
  return res;
}

/**
 * 刷新accessToken
 */
export async function refreshTokenApi() {
  return baseRequestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', {
    withCredentials: true,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/auth/logout', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 * 从 /admin/auth/me 的 permissions 字段获取（不使用单独的 /auth/codes 接口）
 */
export async function getAccessCodesApi() {
  try {
    // 从 /admin/auth/me 获取权限列表
    const me = await getAdminMe();
    return me.permissions || [];
  } catch {
    // 如果获取失败，返回空数组
    return [];
  }
}

/**
 * 修改密码接口参数
 */
export interface ChangePasswordParams {
  oldPassword: string;
  newPassword: string;
}

/**
 * 修改当前管理员密码
 */
export async function changePasswordApi(data: ChangePasswordParams) {
  return requestClient.post<{ success: boolean }>(
    '/admin/auth/change-password',
    {
      oldPassword: data.oldPassword,
      newPassword: data.newPassword,
    },
    {
      responseReturn: 'body',
    },
  );
}
