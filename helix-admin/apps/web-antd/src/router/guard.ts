import type { Router } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { startProgress, stopProgress } from '@vben/utils';

import { accessRoutes, coreRouteNames } from '#/router/routes';
import { useAuthStore } from '#/store';

import { generateAccess } from './access';

/**
 * 通用守卫配置
 * @param router
 */
function setupCommonGuard(router: Router) {
  // 记录已经加载的页面
  const loadedPaths = new Set<string>();

  router.beforeEach((to) => {
    to.meta.loaded = loadedPaths.has(to.path);

    // 页面加载进度条
    if (!to.meta.loaded && preferences.transition.progress) {
      startProgress();
    }
    return true;
  });

  router.afterEach((to) => {
    // 记录页面是否加载,如果已经加载，后续的页面切换动画等效果不在重复执行

    loadedPaths.add(to.path);

    // 关闭页面加载进度条
    if (preferences.transition.progress) {
      stopProgress();
    }
  });
}

/**
 * 权限访问守卫配置
 * @param router
 */
function setupAccessGuard(router: Router) {
  router.beforeEach(async (to, from) => {
    const accessStore = useAccessStore();
    const userStore = useUserStore();
    const authStore = useAuthStore();

    // ================================
    // ✅ 系统配置页：管理员 ID 白名单（不依赖 permissions）
    // 说明：
    // - 仅当路由 meta.adminAllowlistOnly === true 时生效
    // - 白名单通过后直接放行，并跳过后续 permissions / 动态路由逻辑
    // ================================
    const SYSTEM_CONFIG_ALLOWLIST_IDS = new Set<number>([
      2,4 // TODO: 填入允许访问系统配置的 adminId
    ]);

    if (to.meta?.adminAllowlistOnly === true) {
      const adminId = Number(authStore.currentUser?.id);

      // 未获取到当前管理员信息（通常是未登录或 /me 尚未完成）
      if (!Number.isFinite(adminId)) {
        return {
          path: LOGIN_PATH,
          query: { redirect: encodeURIComponent(to.fullPath) },
          replace: true,
        };
      }

      // 不在白名单，直接 403
      if (!SYSTEM_CONFIG_ALLOWLIST_IDS.has(adminId)) {
        return { path: '/403', replace: true };
      }

      // 白名单通过，继续后续登录态 / 动态路由流程
      // 不 return，避免跳过 fetchMe / 动态路由生成
      ;
    }

    // 基本路由，这些路由不需要进入权限拦截
    if (coreRouteNames.includes(to.name as string)) {
      if (to.path === LOGIN_PATH && accessStore.accessToken) {
        return decodeURIComponent(
          (to.query?.redirect as string) ||
            userStore.userInfo?.homePath ||
            preferences.app.defaultHomePath,
        );
      }
      return true;
    }

    // accessToken 检查
    if (!accessStore.accessToken) {
      // 明确声明忽略权限访问权限，则可以访问
      if (to.meta.ignoreAccess) {
        return true;
      }

      // 没有访问权限，跳转登录页面
      if (to.fullPath !== LOGIN_PATH) {
        return {
          path: LOGIN_PATH,
          // 如不需要，直接删除 query
          query:
            to.fullPath === preferences.app.defaultHomePath
              ? {}
              : { redirect: encodeURIComponent(to.fullPath) },
          // 携带当前跳转的页面，登录后重新跳转该页面
          replace: true,
        };
      }
      return to;
    }

    // 如果有 token 但 currentUser 或 permissions 为空，自动拉取 /me
    if (accessStore.accessToken && (!authStore.currentUser || !authStore.permissions.length)) {
      try {
        await authStore.fetchMe();
      } catch (error) {
        // 401 错误由 request 拦截器处理，这里不处理，避免死循环
        // 如果 fetchMe 失败，继续执行后续逻辑
      }
    }

    // 是否已经生成过动态路由
    if (accessStore.isAccessChecked) {
      return true;
    }

    // 生成路由表
    // 当前登录用户拥有的角色标识列表
    // 优先使用 currentUser，如果没有则尝试 fetchUserInfo
    let userInfo = userStore.userInfo;
    if (!userInfo && authStore.currentUser) {
      userInfo = {
        ...authStore.currentUser,
        realName: authStore.currentUser.name,
      } as any;
    }
    if (!userInfo) {
      userInfo = await authStore.fetchUserInfo().catch(() => null);
    }
    const userRoles = userInfo?.roles ?? authStore.currentUser?.roles ?? [];

    // 生成菜单和路由
    const { accessibleMenus, accessibleRoutes } = await generateAccess({
      roles: userRoles,
      router,
      // 则会在菜单中显示，但是访问会被重定向到403
      routes: accessRoutes,
    });

    // 保存菜单信息和路由信息
    accessStore.setAccessMenus(accessibleMenus);
    accessStore.setAccessRoutes(accessibleRoutes);
    accessStore.setIsAccessChecked(true);
    const redirectPath = (from.query.redirect ??
      (to.path === preferences.app.defaultHomePath
        ? userInfo.homePath || preferences.app.defaultHomePath
        : to.fullPath)) as string;

    return {
      ...router.resolve(decodeURIComponent(redirectPath)),
      replace: true,
    };
  });
}

/**
 * 项目守卫配置
 * @param router
 */
function createRouterGuard(router: Router) {
  /** 通用 */
  setupCommonGuard(router);
  /** 权限访问 */
  setupAccessGuard(router);
}

export { createRouterGuard };
