import type { Recordable, UserInfo } from '@vben/types';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { resetAllStores, useAccessStore, useUserStore } from '@vben/stores';

import { notification } from 'ant-design-vue';
import { defineStore } from 'pinia';

import { getAccessCodesApi, getAdminMe, getUserInfoApi, loginApi, logoutApi } from '#/api';
import { $t } from '#/locales';

export interface AdminUser {
  id: string | number;
  name: string;
  roles: string[];
}

export const useAuthStore = defineStore('auth', () => {
  const accessStore = useAccessStore();
  const userStore = useUserStore();
  const router = useRouter();

  const loginLoading = ref(false);
  const currentUser = ref<AdminUser | undefined>(undefined);
  const permissions = ref<string[]>([]);

  /**
   * 异步处理登录操作
   * Asynchronously handle the login process
   * @param params 登录表单数据
   */
  async function authLogin(
    params: Recordable<any>,
    onSuccess?: () => Promise<void> | void,
  ) {
    // 异步处理用户登录操作并获取 accessToken
    let userInfo: null | UserInfo = null;
    try {
      loginLoading.value = true;
      const { accessToken } = await loginApi(params);

      // 如果成功获取到 accessToken
      if (accessToken) {
        accessStore.setAccessToken(accessToken);

        // 登录成功后只保存 token，不主动拉取 /me
        // /me 的拉取统一由 router guard 驱动，避免重复请求
        // 尝试获取用户信息（如果后端有 /user/info 接口）
        try {
          userInfo = await fetchUserInfo();
          userStore.setUserInfo(userInfo);
        } catch {
          // 如果 getUserInfoApi 失败，忽略（由 guard 中的 fetchMe 处理）
        }

        // 获取权限码（如果后端有 /auth/codes 接口）
        const accessCodes = await getAccessCodesApi().catch(() => []);
        accessStore.setAccessCodes(accessCodes);

        if (accessStore.loginExpired) {
          accessStore.setLoginExpired(false);
        } else {
          onSuccess
            ? await onSuccess?.()
            : await router.push(
                userInfo.homePath || preferences.app.defaultHomePath,
              );
        }

        if (userInfo?.realName) {
          notification.success({
            description: `${$t('authentication.loginSuccessDesc')}:${userInfo?.realName}`,
            duration: 3,
            message: $t('authentication.loginSuccess'),
          });
        }
      }
    } finally {
      loginLoading.value = false;
    }

    return {
      userInfo,
    };
  }

  async function logout(redirect: boolean = true) {
    try {
      await logoutApi();
    } catch {
      // 不做任何处理
    }
    // 清空用户信息和权限
    currentUser.value = undefined;
    permissions.value = [];
    resetAllStores();
    accessStore.setLoginExpired(false);

    // 回登录页带上当前路由地址
    await router.replace({
      path: LOGIN_PATH,
      query: redirect
        ? {
            redirect: encodeURIComponent(router.currentRoute.value.fullPath),
          }
        : {},
    });
  }

  async function fetchUserInfo() {
    let userInfo: null | UserInfo = null;
    userInfo = await getUserInfoApi();
    userStore.setUserInfo(userInfo);
    return userInfo;
  }

  /**
   * 获取当前管理员信息（/admin/auth/me）
   * 如果已有 currentUser 和 permissions，则跳过
   */
  async function fetchMe() {
    // 如果没有 token，不调用
    if (!accessStore.accessToken) {
      return null;
    }

    // 如果已有数据，跳过
    if (currentUser.value && permissions.value.length > 0) {
      return {
        id: currentUser.value.id,
        name: currentUser.value.name,
        roles: currentUser.value.roles,
        permissions: permissions.value,
      };
    }

    try {
      const meData = await getAdminMe();
      if (meData) {
        currentUser.value = {
          id: meData.id,
          name: meData.name,
          roles: meData.roles || [],
        };
        permissions.value = meData.permissions || [];
        
        // 同步到 userStore（兼容 Vben 现有逻辑）
        userStore.setUserInfo({
          ...currentUser.value,
          realName: meData.name,
        } as UserInfo);
      }
      return meData;
    } catch (error) {
      // 401 错误由 request 拦截器处理，这里只记录但不抛出
      // 避免死循环
      console.warn('Failed to fetch admin me:', error);
      return null;
    }
  }

  function $reset() {
    loginLoading.value = false;
    currentUser.value = undefined;
    permissions.value = [];
  }

  return {
    $reset,
    authLogin,
    fetchUserInfo,
    fetchMe,
    loginLoading,
    logout,
    currentUser,
    permissions,
  };
});
