<script lang="ts" setup>
import type { AdminUsersApi } from '#/api';

import { h, onMounted, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  Input,
  Modal,
  Select,
  SelectOption,
  Space,
  Switch,
  Table,
  Tag,
  message,
} from 'ant-design-vue';

import {
  createAdminUser,
  getAdminRoles,
  getAdminUsers,
  resetAdminPassword,
  setAdminEnabled,
  setAdminUserRoles,
} from '#/api';

defineOptions({ name: 'AdminUsersList' });

// 查询表单
const searchForm = ref({
  keyword: '',
});

// 表格数据
const loading = ref(false);
const tableData = ref<AdminUsersApi.AdminUser[]>([]);
const total = ref(0);
const pagination = ref({
  current: 1,
  pageSize: 10,
});

// 弹窗状态
const createModalVisible = ref(false);
const resetPasswordModalVisible = ref(false);
const currentUser = ref<AdminUsersApi.AdminUser | null>(null);

// 表单数据
const createForm = ref({
  username: '',
  password: '',
  roles: [] as string[],
});
const resetPasswordForm = ref({
  newPassword: '',
});

// 角色配置弹窗
const rolesModalVisible = ref(false);
const rolesOptions = ref<string[]>([]);
const selectedRoles = ref<string[]>([]);
const rolesLoading = ref(false);

// 提交状态
const submitting = ref(false);

// 表格列
const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: '用户名',
    dataIndex: 'username',
    key: 'username',
  },
  {
    title: '角色',
    dataIndex: 'roles',
    key: 'roles',
    customRender: ({ record }: { record: AdminUsersApi.AdminUser }) => {
      return h(
        'div',
        { style: 'display: flex; gap: 8px; flex-wrap: wrap;' },
        record.roles.map((role) =>
          h(Tag, { key: role, color: 'blue' }, () => role),
        ),
      );
    },
  },
  {
    title: '状态',
    dataIndex: 'isEnabled',
    key: 'isEnabled',
    width: 100,
    customRender: ({ record }: { record: AdminUsersApi.AdminUser }) => {
      return h(Switch, {
        checked: record.isEnabled,
        onChange: (checked: unknown) =>
          handleToggleEnabled(record, checked === true),
      });
    },
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ record }: { record: AdminUsersApi.AdminUser }) =>
      record.createdAt || '-',
  },
  {
    title: '操作',
    key: 'action',
    width: 200,
    customRender: ({ record }: { record: AdminUsersApi.AdminUser }) => {
      return h(Space, null, {
        default: () => [
          h(
            Button,
            {
              type: 'link',
              size: 'small',
              onClick: () => handleConfigRoles(record),
            },
            () => '配置角色',
          ),
          h(
            Button,
            {
              type: 'link',
              size: 'small',
              onClick: () => handleResetPassword(record),
            },
            () => '重置密码',
          ),
        ],
      });
    },
  },
];

// 加载列表
async function loadAdminUsers() {
  loading.value = true;
  try {
    const params: AdminUsersApi.ListParams = {
      keyword: searchForm.value.keyword || undefined,
      page: pagination.value.current,
      pageSize: pagination.value.pageSize,
    };
    const res = await getAdminUsers(params);
    tableData.value = res.list || [];
    total.value = res.total ?? res.list.length ?? 0;
  } catch (error: any) {
    handleApiError(error, '加载管理员列表失败');
  } finally {
    loading.value = false;
  }
}

// 搜索
function handleSearch() {
  pagination.value.current = 1;
  loadAdminUsers();
}

// 重置
function handleReset() {
  searchForm.value = {
    keyword: '',
  };
  pagination.value.current = 1;
  loadAdminUsers();
}

// 分页变化
function handleTableChange(pag: any) {
  pagination.value.current = pag.current ?? 1;
  pagination.value.pageSize = pag.pageSize ?? 10;
  loadAdminUsers();
}

// 打开创建弹窗
function openCreateModal() {
  createForm.value = {
    username: '',
    password: '',
    roles: [],
  };
  createModalVisible.value = true;
}

// 提交创建
async function handleCreateSubmit() {
  if (!createForm.value.username.trim()) {
    message.error('请输入用户名');
    return;
  }

  if (!createForm.value.password.trim()) {
    message.error('请输入密码');
    return;
  }

  try {
    submitting.value = true;
    await createAdminUser({
      username: createForm.value.username,
      password: createForm.value.password,
      roles: createForm.value.roles,
    });
    message.success('创建成功');
    createModalVisible.value = false;
    loadAdminUsers();
  } catch (error: any) {
    handleApiError(error, '创建失败');
  } finally {
    submitting.value = false;
  }
}

// 切换启用状态
async function handleToggleEnabled(
  user: AdminUsersApi.AdminUser,
  isEnabled: boolean,
) {
  Modal.confirm({
    title: '确认操作',
    content: `确认${isEnabled ? '启用' : '禁用'}管理员 ${user.username} ？`,
    onOk: async () => {
      try {
        await setAdminEnabled(user.id, isEnabled);
        message.success('操作成功');
        loadAdminUsers();
      } catch (error: any) {
        handleApiError(error, '操作失败');
      }
    },
  });
}

// 配置角色
function handleConfigRoles(user: AdminUsersApi.AdminUser) {
  currentUser.value = user;
  rolesModalVisible.value = true;
  rolesLoading.value = true;
  selectedRoles.value = [...(user.roles || [])];

  getAdminRoles()
    .then((roles) => {
      rolesOptions.value = roles.map((role) => role.name);
    })
    .catch((error: any) => {
      handleApiError(error, '加载角色列表失败');
      rolesModalVisible.value = false;
    })
    .finally(() => {
      rolesLoading.value = false;
    });
}

// 保存角色
async function handleRolesSubmit() {
  if (!currentUser.value) return;
  try {
    submitting.value = true;
    await setAdminUserRoles(currentUser.value.id, selectedRoles.value);
    message.success('保存角色成功');
    rolesModalVisible.value = false;
    loadAdminUsers();
  } catch (error: any) {
    handleApiError(error, '保存角色失败');
  } finally {
    submitting.value = false;
  }
}

// 重置密码
function handleResetPassword(user: AdminUsersApi.AdminUser) {
  currentUser.value = user;
  resetPasswordForm.value = {
    newPassword: '',
  };
  resetPasswordModalVisible.value = true;
}

// 提交重置密码
async function handleResetPasswordSubmit() {
  if (!currentUser.value) return;

  if (!resetPasswordForm.value.newPassword.trim()) {
    message.error('请输入新密码');
    return;
  }

  if (resetPasswordForm.value.newPassword.length < 6) {
    message.error('密码长度至少 6 位');
    return;
  }

  Modal.confirm({
    title: '确认重置密码',
    content: `确认重置管理员 ${currentUser.value.username} 的密码？`,
    onOk: async () => {
      try {
        submitting.value = true;
        await resetAdminPassword(
          currentUser.value!.id,
          resetPasswordForm.value.newPassword,
        );
        message.success('重置密码成功');
        resetPasswordModalVisible.value = false;
      } catch (error: any) {
        handleApiError(error, '重置密码失败');
      } finally {
        submitting.value = false;
      }
    },
  });
}

function handleApiError(error: any, fallbackMsg: string) {
  const status = error?.response?.status;
  const code = error?.response?.data?.error;
  const msg =
    error?.response?.data?.message || error?.message || fallbackMsg;

  if (status === 403) {
    if (code === 'USER_DISABLED') {
      message.error('账号已禁用，请重新登录');
      // 交由全局拦截器处理跳转，这里只提示
      return;
    }
    if (code === 'FORBIDDEN') {
      message.error('无权限（admin.manage）');
      return;
    }
  }

  message.error(msg);
}

onMounted(() => {
  loadAdminUsers();
});
</script>

<template>
  <div class="p-5">
    <!-- 查询表单 -->
    <Card class="mb-4">
      <Form :model="searchForm" layout="inline">
        <Form.Item label="关键词">
          <Input
            v-model:value="searchForm.keyword"
            placeholder="用户名或ID"
            style="width: 200px"
            @press-enter="handleSearch"
          />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" @click="handleSearch">搜索</Button>
            <Button @click="handleReset">重置</Button>
            <Button type="primary" @click="openCreateModal">新建管理员</Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>

    <!-- 表格 -->
    <Card>
      <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: total,
        showTotal: (total) => `共 ${total} 条`,
      }"
      @change="handleTableChange"
      />

      <!-- 创建弹窗 -->
    <Modal
      v-model:open="createModalVisible"
      title="新建管理员"
      :confirm-loading="submitting"
      @ok="handleCreateSubmit"
    >
      <Form :model="createForm" layout="vertical">
        <Form.Item label="用户名" required>
          <Input v-model:value="createForm.username" placeholder="请输入用户名" />
        </Form.Item>
        <Form.Item label="密码" required>
          <Input
            v-model:value="createForm.password"
            type="password"
            placeholder="请输入密码"
          />
        </Form.Item>
        <Form.Item label="角色">
          <p class="text-gray-500 text-sm">
            创建后可在"配置角色"中设置角色
          </p>
        </Form.Item>
      </Form>
    </Modal>

    <!-- 重置密码弹窗 -->
    <Modal
      v-model:open="resetPasswordModalVisible"
      title="重置密码"
      :confirm-loading="submitting"
      @ok="handleResetPasswordSubmit"
    >
      <Form :model="resetPasswordForm" layout="vertical">
        <Form.Item label="新密码" required>
          <Input
            v-model:value="resetPasswordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6位）"
          />
        </Form.Item>
      </Form>
    </Modal>

    <!-- 配置角色弹窗 -->
    <Modal
      v-model:open="rolesModalVisible"
      title="配置角色"
      :confirm-loading="submitting"
      :ok-button-props="{ disabled: rolesLoading }"
      @ok="handleRolesSubmit"
    >
      <Form layout="vertical">
        <Form.Item label="角色">
          <Select
            v-model:value="selectedRoles"
            mode="multiple"
            placeholder="请选择角色"
            :loading="rolesLoading"
            style="width: 100%"
          >
            <SelectOption
              v-for="role in rolesOptions"
              :key="role"
              :value="role"
            >
              {{ role }}
            </SelectOption>
          </Select>
        </Form.Item>
      </Form>
    </Modal>
    </Card>
  </div>
</template>

