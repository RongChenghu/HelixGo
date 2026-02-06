# Vben 登录流程对接文档

## 修改文件清单

### 1. 环境变量配置
- `admin/apps/web-antd/.env.development` - 添加 `VITE_ADMIN_API_BASE_URL=http://localhost:3000`

### 2. API 请求配置
- `admin/apps/web-antd/src/api/request.ts` - 修改 baseURL 和 401 处理

### 3. API 接口
- `admin/apps/web-antd/src/api/core/auth.ts` - 对接登录和 /me 接口
- `admin/apps/web-antd/src/api/core/index.ts` - 导出 getAdminMe

### 4. Store
- `admin/apps/web-antd/src/store/auth.ts` - 添加 currentUser、permissions、fetchMe

### 5. 路由守卫
- `admin/apps/web-antd/src/router/guard.ts` - 自动拉取 /me

### 6. 权限判断
- `admin/apps/web-antd/src/router/access.ts` - 实现 can() 函数

### 7. 示例页面
- `admin/apps/web-antd/src/views/dashboard/workspace/index.vue` - 添加权限验证示例

---

## 详细 Diff

### 1. admin/apps/web-antd/src/api/request.ts

**修改点：**
- 第 6 行：添加 `LOGIN_PATH` 导入
- 第 23-24 行：修改 baseURL 为 `VITE_ADMIN_API_BASE_URL`
- 第 32-60 行：修改 `doReAuthenticate` 函数
  - 清空 token 和用户信息
  - 使用 router.replace 跳转登录页（不 reload）
- 第 110-126 行：修改错误处理，跳过 401（已由 authenticateResponseInterceptor 处理）

**关键代码：**
```typescript
// 使用 VITE_ADMIN_API_BASE_URL 作为 baseURL
const apiURL = import.meta.env.VITE_ADMIN_API_BASE_URL || useAppConfig(import.meta.env, import.meta.env.PROD).apiURL;

// 401 处理：清空 token 和用户信息，跳转登录页
async function doReAuthenticate() {
  const accessStore = useAccessStore();
  const authStore = useAuthStore();
  
  accessStore.setAccessToken(null);
  if (authStore.currentUser || authStore.permissions.length > 0) {
    authStore.currentUser = undefined;
    authStore.permissions = [];
  }
  
  const router = (await import('vue-router')).useRouter();
  router().replace({ path: LOGIN_PATH });
}
```

---

### 2. admin/apps/web-antd/src/api/core/auth.ts

**修改点：**
- 第 21-44 行：修改 `loginApi` 函数
  - 对接 `POST /admin/auth/login`
  - 使用 `responseReturn: 'body'` 获取响应体
  - 提取 token 并转换为 `{ accessToken }` 格式
- 第 46-60 行：新增 `getAdminMe` 函数
  - 对接 `GET /admin/auth/me`
  - 返回用户信息和权限列表
- 第 80-91 行：修改 `getAccessCodesApi`，添加容错处理

**关键代码：**
```typescript
// 登录接口
export async function loginApi(data: AuthApi.LoginParams) {
  const res = await requestClient.post<{ token: string }>(
    '/admin/auth/login',
    { username: data.username, password: data.password },
    { responseReturn: 'body' }
  );
  return { accessToken: (res as any)?.token ?? res };
}

// 获取用户信息
export async function getAdminMe() {
  const res = await requestClient.get<{
    id: string | number;
    name: string;
    roles: string[];
    permissions: string[];
  }>('/admin/auth/me', { responseReturn: 'body' });
  return res;
}
```

---

### 3. admin/apps/web-antd/src/store/auth.ts

**修改点：**
- 第 13 行：添加 `getAdminMe` 导入
- 第 16-20 行：新增 `AdminUser` 接口
- 第 27-29 行：新增 state：`currentUser` 和 `permissions`
- 第 50-83 行：修改 `authLogin`，优先使用 `/admin/auth/me` 数据
- 第 112-133 行：修改 `logout`，清空 currentUser 和 permissions
- 第 142-185 行：新增 `fetchMe` 函数
  - 检查 token 是否存在
  - 如果已有数据则跳过
  - 调用 `/admin/auth/me` 并存储结果
- 第 187-191 行：修改 `$reset`，清空新增的 state
- 第 193-202 行：导出新增的 state 和函数

**关键代码：**
```typescript
const currentUser = ref<AdminUser | undefined>(undefined);
const permissions = ref<string[]>([]);

async function fetchMe() {
  if (!accessStore.accessToken) return null;
  if (currentUser.value && permissions.value.length > 0) {
    return { ...currentUser.value, permissions: permissions.value };
  }
  
  try {
    const meData = await getAdminMe();
    currentUser.value = { id: meData.id, name: meData.name, roles: meData.roles };
    permissions.value = meData.permissions || [];
    return meData;
  } catch (error) {
    console.warn('Failed to fetch admin me:', error);
    return null;
  }
}
```

---

### 4. admin/apps/web-antd/src/router/guard.ts

**修改点：**
- 第 88-96 行：在路由守卫中添加自动拉取 /me 的逻辑
  - 检查 token 是否存在
  - 检查 currentUser 或 permissions 是否为空
  - 调用 `fetchMe()` 获取用户信息
  - 错误处理：401 由 request 拦截器处理，避免死循环
- 第 103-116 行：修改用户信息获取逻辑
  - 优先使用 `currentUser`
  - 兼容 `userStore.userInfo`

**关键代码：**
```typescript
// 如果有 token 但 currentUser 或 permissions 为空，自动拉取 /me
if (accessStore.accessToken && (!authStore.currentUser || !authStore.permissions.length)) {
  try {
    await authStore.fetchMe();
  } catch (error) {
    // 401 错误由 request 拦截器处理，避免死循环
  }
}
```

---

### 5. admin/apps/web-antd/src/router/access.ts

**修改点：**
- 第 14 行：添加 `useAuthStore` 导入
- 第 43-66 行：新增 `can()` 函数
  - 从 `authStore.permissions` 读取权限数组
  - 支持单个权限字符串：`can('wallet.credit')`
  - 支持权限数组：`can(['a', 'b'])` - 任意一个为 true 即返回 true
  - 如果 permission 为空，返回 false

**关键代码：**
```typescript
export function can(permission: string | string[]): boolean {
  const authStore = useAuthStore();
  const permissions = authStore.permissions || [];
  
  if (!permission) return false;
  
  if (Array.isArray(permission)) {
    if (permission.length === 0) return false;
    return permission.some((p) => permissions.includes(p));
  }
  
  return permissions.includes(permission);
}
```

---

### 6. admin/apps/web-antd/src/views/dashboard/workspace/index.vue

**修改点：**
- 第 25 行：导入 `can` 函数
- 第 28 行：创建 `access` 对象
- 第 247-251 行：添加权限验证示例按钮

**关键代码：**
```vue
<script>
import { can } from '#/router/access';
const access = { can };
</script>

<template>
  <div class="mt-4 flex gap-2">
    <a-button v-if="access.can('wallet.credit')" type="primary">
      充值（需要 wallet.credit 权限）
    </a-button>
    <a-button v-if="access.can('withdraw.approve')" type="default">
      审核提现（需要 withdraw.approve 权限）
    </a-button>
    <a-button v-if="access.can(['wallet.credit', 'wallet.debit'])" type="dashed">
      钱包操作（需要 wallet.credit 或 wallet.debit）
    </a-button>
  </div>
</template>
```

---

## 自测说明

### 1. 环境配置

在 `admin/apps/web-antd/.env.development` 中确认：
```bash
VITE_ADMIN_API_BASE_URL=http://localhost:3000
```

### 2. 后端服务

确保后端服务运行在 `http://localhost:3000`，并已配置：
- `ADMIN_BOOTSTRAP_USERNAME=admin`
- `ADMIN_BOOTSTRAP_PASSWORD=<你的密码>`
- `ADMIN_JWT_SECRET=<你的密钥>`

### 3. 启动前端

```bash
cd admin
pnpm dev:antd
```

### 4. 测试登录

1. 打开浏览器访问 `http://localhost:5666`
2. 输入用户名：`admin`
3. 输入密码：`<你的密码>`
4. 点击登录

**预期结果：**
- 登录成功，跳转到工作台页面
- 浏览器 Network 标签中可以看到：
  - `POST /admin/auth/login` 返回 `{ token: "..." }`
  - `GET /admin/auth/me` 返回用户信息和权限列表

### 5. 测试权限验证

1. 登录成功后，进入工作台页面
2. 查看页面上的权限按钮：
   - "充值" 按钮应该显示（有 `wallet.credit` 权限）
   - "审核提现" 按钮应该显示（有 `withdraw.approve` 权限）
   - "钱包操作" 按钮应该显示（有 `wallet.credit` 或 `wallet.debit` 权限）

### 6. 测试刷新页面

1. 登录成功后，刷新页面（F5）
2. 浏览器 Network 标签中应该看到：
   - `GET /admin/auth/me` 请求（自动拉取用户信息）
   - 请求成功，页面正常显示

### 7. 测试 401 处理

1. 在浏览器控制台执行：`localStorage.clear()`
2. 或者修改后端返回 401
3. 刷新页面

**预期结果：**
- 自动跳转到登录页
- 不显示重复的错误提示
- 不出现死循环

---

## 功能验证清单

- [x] 登录接口对接：`POST /admin/auth/login` 返回 token
- [x] Token 存储：使用 Vben 现有的 `accessStore.setAccessToken()`
- [x] 自动注入 Bearer token：请求拦截器已配置
- [x] 获取用户信息：`GET /admin/auth/me` 返回 roles + permissions
- [x] 刷新页面自动拉取：路由守卫中自动调用 `fetchMe()`
- [x] 权限判断：`can()` 函数可用
- [x] 401 处理：清 token、跳登录、不 reload、不重复 message
- [x] 无 token 不调用 /me：`fetchMe()` 中有检查

---

## 使用示例

### 在组件中使用权限判断

```vue
<script setup lang="ts">
import { can } from '#/router/access';

const access = { can };
</script>

<template>
  <div>
    <a-button v-if="access.can('wallet.credit')">充值</a-button>
    <a-button v-if="access.can(['withdraw.approve', 'withdraw.pay')]">
      提现操作
    </a-button>
  </div>
</template>
```

### 在脚本中使用权限判断

```typescript
import { can } from '#/router/access';

if (can('wallet.credit')) {
  // 执行充值操作
}

if (can(['wallet.credit', 'wallet.debit'])) {
  // 执行钱包操作（任意一个权限即可）
}
```

---

## 注意事项

1. **响应格式**：后端直接返回 `{ token }` 和用户信息对象，没有 `code/data` 包装，所以使用了 `responseReturn: 'body'`
2. **401 处理**：已确保不重复处理，401 错误由 `authenticateResponseInterceptor` 统一处理
3. **权限数组**：`can(['a', 'b'])` 表示任意一个权限为 true 即返回 true（OR 逻辑）
4. **刷新页面**：路由守卫会自动检查并拉取用户信息，无需手动调用
