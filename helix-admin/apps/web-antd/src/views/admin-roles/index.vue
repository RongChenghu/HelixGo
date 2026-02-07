<script lang="ts" setup>
import type { AdminRolesApi } from '#/api';
import type { PermissionItem } from '#/api';

import { onMounted, ref } from 'vue';

import {
  Button,
  Card,
  Form,
  FormItem,
  Input,
  Modal,
  Select,
  SelectOption,
  Space,
  Table,
  Tag,
  message,
} from 'ant-design-vue';

import {
  createRole,
  deleteRole,
  getAdminRoles,
  getPermissions,
  updateRole,
} from '#/api';

defineOptions({ name: 'AdminRolesList' });

const loading = ref(false);
const tableData = ref<AdminRolesApi.Role[]>([]);

const createModalVisible = ref(false);
const editModalVisible = ref(false);
const currentRole = ref<AdminRolesApi.Role | null>(null);
const permissionsOptions = ref<PermissionItem[]>([]);
const submitting = ref(false);

const createForm = ref({
  name: '',
  description: '',
  perms: [] as string[],
});
const editForm = ref({
  description: '',
  perms: [] as string[],
});

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 80,
  },
  {
    title: '角色名称',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: '描述',
    dataIndex: 'description',
    key: 'description',
  },
  {
    title: '权限代码',
    dataIndex: 'perms',
    key: 'perms',
  },
  {
    title: '操作',
    key: 'action',
    width: 180,
  },
];

async function loadRoles() {
  loading.value = true;
  try {
    const roles = await getAdminRoles();
    tableData.value = Array.isArray(roles) ? roles : [];
  } catch (error: any) {
    handleApiError(error, '加载角色列表失败');
  } finally {
    loading.value = false;
  }
}

async function loadPermissions() {
  try {
    permissionsOptions.value = await getPermissions();
  } catch (error: any) {
    handleApiError(error, '加载权限字典失败');
  }
}

function handleApiError(error: any, fallbackMsg: string) {
  const status = error?.response?.status;
  const code = error?.response?.data?.error;
  const msg =
    error?.response?.data?.message || error?.message || fallbackMsg;
  if (status === 403 && code === 'FORBIDDEN') {
    message.error('无权限（需 admin.manage 或 admin.role.write）');
    return;
  }
  message.error(msg);
}

function openCreateModal() {
  createForm.value = { name: '', description: '', perms: [] };
  createModalVisible.value = true;
}

async function openEditModal(record: AdminRolesApi.Role) {
  currentRole.value = record;
  editForm.value = {
    description: record.description ?? '',
    perms: record.perms ?? [],
  };
  editModalVisible.value = true;
}

async function handleCreateSubmit() {
  if (!createForm.value.name.trim()) {
    message.error('请输入角色名称');
    return;
  }
  submitting.value = true;
  try {
    await createRole({
      name: createForm.value.name.trim(),
      description: createForm.value.description?.trim() || undefined,
      perms:
        createForm.value.perms?.length > 0 ? createForm.value.perms : undefined,
    });
    message.success('创建成功');
    createModalVisible.value = false;
    loadRoles();
  } catch (error: any) {
    handleApiError(error, '创建角色失败');
  } finally {
    submitting.value = false;
  }
}

async function handleEditSubmit() {
  if (!currentRole.value) return;
  submitting.value = true;
  try {
    await updateRole(currentRole.value.id, {
      description: editForm.value.description?.trim() || undefined,
      perms:
        editForm.value.perms?.length > 0 ? editForm.value.perms : undefined,
    });
    message.success('更新成功');
    editModalVisible.value = false;
    loadRoles();
  } catch (error: any) {
    handleApiError(error, '更新角色失败');
  } finally {
    submitting.value = false;
  }
}

function handleDelete(record: AdminRolesApi.Role) {
  if (record.name === 'admin') {
    message.error('不可删除系统内置角色 admin');
    return;
  }
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除角色「${record.name}」吗？删除后该角色下的管理员将失去该角色权限。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    onOk: async () => {
      try {
        await deleteRole(record.id);
        message.success('删除成功');
        loadRoles();
      } catch (error: any) {
        const code = error?.response?.data?.error;
        const msg =
          error?.response?.data?.message ||
          error?.message ||
          '删除角色失败';
        if (code === 'ROLE_PROTECTED') {
          message.error('不可删除系统内置角色');
          return;
        }
        handleApiError(error, msg);
      }
    },
  });
}

onMounted(() => {
  loadRoles();
  loadPermissions();
});
</script>

<template>
  <div class="p-5">
    <Card class="mb-4">
      <Space>
        <Button type="primary" @click="openCreateModal">新建角色</Button>
      </Space>
    </Card>

    <Card>
      <Table
        :columns="columns"
        :data-source="tableData"
        :loading="loading"
        row-key="id"
        :pagination="false"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'perms'">
            <span v-if="!record.perms || !record.perms.length">-</span>
            <template v-else>
              <Tag
                v-for="perm in record.perms"
                :key="perm"
                color="blue"
                class="perm-tag"
              >
                {{ perm }}
              </Tag>
            </template>
          </template>
          <template v-else-if="column.key === 'action'">
            <Space>
              <Button
                type="link"
                size="small"
                @click="openEditModal(record as AdminRolesApi.Role)"
              >
                编辑
              </Button>
              <Button
                v-if="record.name !== 'admin'"
                type="link"
                size="small"
                danger
                @click="handleDelete(record as AdminRolesApi.Role)"
              >
                删除
              </Button>
            </Space>
          </template>
        </template>
      </Table>
    </Card>

    <Modal
      v-model:open="createModalVisible"
      title="新建角色"
      :confirm-loading="submitting"
      @ok="handleCreateSubmit"
    >
      <Form :model="createForm" layout="vertical">
        <FormItem label="角色名称" required>
          <Input
            v-model:value="createForm.name"
            placeholder="请输入角色名称（唯一）"
          />
        </FormItem>
        <FormItem label="描述">
          <Input
            v-model:value="createForm.description"
            placeholder="选填"
          />
        </FormItem>
        <FormItem label="权限">
          <Select
            v-model:value="createForm.perms"
            mode="multiple"
            placeholder="请选择权限"
            style="width: 100%"
          >
            <SelectOption
              v-for="p in permissionsOptions"
              :key="p.code"
              :value="p.code"
            >
              {{ p.code }} - {{ p.description }}
            </SelectOption>
          </Select>
        </FormItem>
      </Form>
    </Modal>

    <Modal
      v-model:open="editModalVisible"
      title="编辑角色"
      :confirm-loading="submitting"
      @ok="handleEditSubmit"
    >
      <Form v-if="currentRole" :model="editForm" layout="vertical">
        <FormItem label="角色名称">
          <Input :value="currentRole.name" disabled />
        </FormItem>
        <FormItem label="描述">
          <Input
            v-model:value="editForm.description"
            placeholder="选填"
          />
        </FormItem>
        <FormItem label="权限">
          <Select
            v-model:value="editForm.perms"
            mode="multiple"
            placeholder="请选择权限"
            style="width: 100%"
          >
            <SelectOption
              v-for="p in permissionsOptions"
              :key="p.code"
              :value="p.code"
            >
              {{ p.code }} - {{ p.description }}
            </SelectOption>
          </Select>
        </FormItem>
      </Form>
    </Modal>
  </div>
</template>

<style scoped>
.perm-tag {
  display: inline-block;
  margin-right: 4px;
  margin-bottom: 4px;
  padding: 2px 6px;
  font-size: 12px;
  background-color: #f0f0f0;
  border-radius: 4px;
}
</style>
