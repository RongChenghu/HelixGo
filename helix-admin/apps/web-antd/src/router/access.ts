import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
} from '@vben/types';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';

import { message } from 'ant-design-vue';

import { getAllMenusApi } from '#/api';
import { BasicLayout, IFrameView } from '#/layouts';
import { $t } from '#/locales';
import { useAuthStore } from '#/store';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

async function generateAccess(options: GenerateMenuAndRoutesOptions) {
  const pageMap: ComponentRecordType = import.meta.glob('../views/**/*.vue');

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  return await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => {
      message.loading({
        content: `${$t('common.loadingMenu')}...`,
        duration: 1.5,
      });
      return await getAllMenusApi();
    },
    // 可以指定没有权限跳转403页面
    forbiddenComponent,
    // 如果 route.meta.menuVisibleWithForbidden = true
    layoutMap,
    pageMap,
  });
}

/**
 * 权限判断函数
 * @param permission 权限字符串或权限数组
 * @returns 是否有权限
 */
function can(permission: string | string[]): boolean {
  const authStore = useAuthStore();
  const permissions = authStore.permissions || [];

  if (!permission) {
    return false;
  }

  // 如果有 * 权限，表示全量权限，直接返回 true
  if (permissions.includes('*')) {
    return true;
  }

  // 如果是数组，任意一个为 true 即返回 true
  if (Array.isArray(permission)) {
    if (permission.length === 0) {
      return false;
    }
    return permission.some((p) => permissions.includes(p));
  }

  // 如果是字符串，直接判断
  return permissions.includes(permission);
}

/**
 * 创建统一的权限访问对象
 * 供页面组件使用，无需手动 import can
 */
export function createAccess() {
  return {
    can,
  };
}

export { generateAccess };
