/**
 * 该文件可自行根据业务逻辑进行调整
 */
import type { RequestClientOptions } from '@vben/request';

import { LOGIN_PATH } from '@vben/constants';
import { useAppConfig } from '@vben/hooks';
import { preferences } from '@vben/preferences';
import {
  authenticateResponseInterceptor,
  defaultResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';
import { useAccessStore } from '@vben/stores';

import { message } from 'ant-design-vue';

import { router } from '#/router';
import { useAuthStore } from '#/store';

import { refreshTokenApi } from './core';

// 使用 VITE_ADMIN_API_BASE_URL 作为 baseURL
const apiURL = import.meta.env.VITE_ADMIN_API_BASE_URL || useAppConfig(import.meta.env, import.meta.env.PROD).apiURL;

function createRequestClient(baseURL: string, options?: RequestClientOptions) {
  const client = new RequestClient({
    ...options,
    baseURL,
  });

  /**
   * 重新认证逻辑
   * 401 错误时清空 token 并跳转登录页
   */
  async function doReAuthenticate() {
    const accessStore = useAccessStore();
    const authStore = useAuthStore();
    
    // 清空 token
    accessStore.setAccessToken(null);
    
    // 清空用户信息和权限
    if (authStore.currentUser || authStore.permissions.length > 0) {
      authStore.currentUser = undefined;
      authStore.permissions = [];
    }
    
    // 跳转登录页（不 reload，不重复 message）
    if (
      preferences.app.loginExpiredMode === 'modal' &&
      accessStore.isAccessChecked
    ) {
      accessStore.setLoginExpired(true);
    } else {
      // 使用 router 实例进行跳转，避免在非组件上下文中使用 useRouter
      await router.replace({ path: LOGIN_PATH });
    }
  }

  /**
   * 刷新token逻辑
   */
  async function doRefreshToken() {
    const accessStore = useAccessStore();
    const resp = await refreshTokenApi();
    const newToken = resp.data;
    accessStore.setAccessToken(newToken);
    return newToken;
  }

  function formatToken(token: null | string) {
    return token ? `Bearer ${token}` : null;
  }

  // 请求头处理
  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();

      config.headers.Authorization = formatToken(accessStore.accessToken);
      config.headers['Accept-Language'] = preferences.app.locale;
      return config;
    },
  });

  // 处理返回的响应数据格式
  // 注意：后端 /admin/auth/login 直接返回 { token }，没有 code/data 包装
  // 如果后端统一返回格式，可以调整这里的配置
  client.addResponseInterceptor(
    defaultResponseInterceptor({
      codeField: 'code',
      dataField: 'data',
      successCode: 0,
    }),
  );

  // token过期的处理
  client.addResponseInterceptor(
    authenticateResponseInterceptor({
      client,
      doReAuthenticate,
      doRefreshToken,
      enableRefreshToken: preferences.app.enableRefreshToken,
      formatToken,
    }),
  );

  // 通用的错误处理,如果没有进入上面的错误处理逻辑，就会进入这里
  // 注意：401 错误已由 authenticateResponseInterceptor 处理，这里不会重复处理
  client.addResponseInterceptor(
    errorMessageResponseInterceptor((msg: string, error) => {
      // 401 错误已由 authenticateResponseInterceptor 处理，跳过
      if (error?.response?.status === 401) {
        return Promise.reject(error);
      }
      
      // 这里可以根据业务进行定制,你可以拿到 error 内的信息进行定制化处理，根据不同的 code 做不同的提示，而不是直接使用 message.error 提示 msg
      // 当前mock接口返回的错误字段是 error 或者 message
      const responseData = error?.response?.data ?? {};
      const errorMessage = responseData?.error ?? responseData?.message ?? '';
      // 如果没有错误信息，则会根据状态码进行提示
      message.error(errorMessage || msg);
    }),
  );

  return client;
}

export const requestClient = createRequestClient(apiURL, {
  responseReturn: 'data',
});

export const baseRequestClient = new RequestClient({ baseURL: apiURL });
