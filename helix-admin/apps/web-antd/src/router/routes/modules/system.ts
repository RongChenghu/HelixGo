import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    meta: {
      icon: 'lucide:settings',
      order: 4,
      title: '系统管理',
    },
    name: 'System',
    path: '/system',
    children: [
      {
        name: 'SystemConfigList',
        path: '/system/configs',
        component: () => import('#/views/system/SystemConfigList.vue'),
        meta: {
          icon: 'lucide:sliders',
          title: '系统配置',
          adminAllowlistOnly: true, // ✅ 白名单管理员专用页面（不走 permissions）
        },
      },
      {
        name: 'AdminUsersList',
        path: '/admin-users',
        component: () => import('#/views/admin-users/AdminUsersList.vue'),
        meta: {
          icon: 'lucide:users',
          title: '管理员管理',
          access: ['admin.manage'],
        },
      },
      {
        name: 'AdminRolesList',
        path: '/admin-roles',
        component: () => import('#/views/admin-roles/index.vue'),
        meta: {
          icon: 'lucide:key',
          title: '角色管理',
          access: ['admin.manage'],
        },
      },
      {
        name: 'AuditLogsList',
        path: '/audit/logs',
        component: () => import('#/views/audit/AuditLogsList.vue'),
        meta: {
          icon: 'lucide:file-text',
          title: '操作审计',
          access: ['admin.manage'],
        },
      },
    ],
  },
];

export default routes;
